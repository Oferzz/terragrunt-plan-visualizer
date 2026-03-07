package server

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

//go:embed static/*
var staticFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Server holds the HTTP server and plan data.
type Server struct {
	port        int
	actualPort  int
	plan        *plan.Plan
	mu          sync.RWMutex
	httpSrv     *http.Server
	applyFunc   func(ctx context.Context, outputFn func(line string)) error
	aiAnalysis  *AIAnalysisData
	wsClients   map[*websocket.Conn]bool
	wsMu        sync.Mutex
}

// AIAnalysisData holds AI analysis results posted via API.
type AIAnalysisData struct {
	Findings        []string `json:"findings"`
	RiskSummary     string   `json:"risk_summary"`
	Recommendations []string `json:"recommendations"`
}

// New creates a new server instance.
func New(port int, p *plan.Plan) *Server {
	return &Server{
		port:      port,
		plan:      p,
		wsClients: make(map[*websocket.Conn]bool),
	}
}

// SetApplyFunc sets the function called when apply is triggered from the UI.
func (s *Server) SetApplyFunc(fn func(ctx context.Context, outputFn func(line string)) error) {
	s.applyFunc = fn
}

// ListenAndServe starts the HTTP server and blocks until context is cancelled.
// onReady is called with the actual bound port after successful port resolution.
func (s *Server) ListenAndServe(ctx context.Context, onReady func(port int)) error {
	resolvedPort, err := resolvePort(s.port, maxPortAttempts, nil)
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}
	s.actualPort = resolvedPort

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.actualPort),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	if onReady != nil {
		onReady(s.actualPort)
	}

	select {
	case <-ctx.Done():
		return s.httpSrv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

// ActualPort returns the port the server bound to.
func (s *Server) ActualPort() int {
	return s.actualPort
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/plan", s.handlePlan)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/api/apply", s.handleApply)
	mux.HandleFunc("/api/ai-analysis", s.handleAIAnalysis)
	mux.HandleFunc("/ws/apply", s.handleWSApply)

	// Serve embedded React app
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub filesystem: %v", err))
	}
	fileServer := http.FileServer(http.FS(sub))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" {
			if _, err := fs.Stat(sub, path[1:]); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		data, err := fs.ReadFile(sub, "index.html")
		if err != nil {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})
}

func (s *Server) handlePlan(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.plan)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (s *Server) handleApply(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if s.applyFunc == nil {
		http.Error(w, `{"error": "apply not available"}`, http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	go func() {
		err := s.applyFunc(context.Background(), func(line string) {
			s.broadcastWS(map[string]interface{}{
				"type":      "stdout",
				"text":      line,
				"timestamp": time.Now().UnixMilli(),
			})
		})

		status := "complete"
		if err != nil {
			status = "error: " + err.Error()
		}
		s.broadcastWS(map[string]interface{}{
			"type":      "status",
			"text":      status,
			"timestamp": time.Now().UnixMilli(),
		})
	}()

	json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

func (s *Server) handleAIAnalysis(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		defer s.mu.RUnlock()
		w.Header().Set("Content-Type", "application/json")
		if s.aiAnalysis != nil {
			json.NewEncoder(w).Encode(s.aiAnalysis)
		} else {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "no analysis available"})
		}
	case http.MethodPost:
		var data AIAnalysisData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.aiAnalysis = &data
		s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleWSApply(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	s.wsMu.Lock()
	s.wsClients[conn] = true
	s.wsMu.Unlock()

	defer func() {
		s.wsMu.Lock()
		delete(s.wsClients, conn)
		s.wsMu.Unlock()
		conn.Close()
	}()

	// Keep connection alive, read messages (for ping/pong)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *Server) broadcastWS(msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	s.wsMu.Lock()
	defer s.wsMu.Unlock()

	for conn := range s.wsClients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			conn.Close()
			delete(s.wsClients, conn)
		}
	}
}

// OpenBrowser opens the default browser to the given URL.
func OpenBrowser(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "linux":
		cmd = "xdg-open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler"}
	default:
		return
	}

	args = append(args, url)
	exec.Command(cmd, args...).Start()
}
