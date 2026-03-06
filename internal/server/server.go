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

	"github.com/ofertzadaka/terragrunt-plan-visualizer/internal/plan"
)

//go:embed static/*
var staticFS embed.FS

// Server holds the HTTP server and plan data.
type Server struct {
	port    int
	plan    *plan.Plan
	mu      sync.RWMutex
	httpSrv *http.Server
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

	// Serve embedded React app
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(fmt.Sprintf("failed to create sub filesystem: %v", err))
	}
	fileServer := http.FileServer(http.FS(sub))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// For SPA routing: serve index.html for non-file paths
		path := r.URL.Path
		if path != "/" {
			// Try to serve the file directly
			if _, err := fs.Stat(sub, path[1:]); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		// Serve index.html for root and SPA routes
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
