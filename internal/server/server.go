package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

// Server holds the HTTP server and plan data.
type Server struct {
	port     int
	plan     *plan.Plan
	mu       sync.RWMutex
	httpSrv  *http.Server
}

// New creates a new server instance.
func New(port int, p *plan.Plan) *Server {
	return &Server{
		port: port,
		plan: p,
	}
}

// ListenAndServe starts the HTTP server and blocks until context is cancelled.
func (s *Server) ListenAndServe(ctx context.Context) error {
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpSrv = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		return s.httpSrv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func (s *Server) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/plan", s.handlePlan)
	mux.HandleFunc("/api/status", s.handleStatus)
	mux.HandleFunc("/", s.handleIndex)
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

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(indexHTML)
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

// Placeholder HTML - will be replaced by embedded React app in Phase 2
var indexHTML = []byte(`<!DOCTYPE html>
<html><head><title>tgviz - loading...</title></head>
<body><p>Web UI placeholder. React app will be embedded here.</p></body></html>
`)
