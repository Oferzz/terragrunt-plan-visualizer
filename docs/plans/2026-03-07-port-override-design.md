# Port Override Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** When the web server port is busy, automatically kill previous tgviz instances or fall back to the next available port.

**Architecture:** Add OS-specific port detection via build-tagged files (`port_unix.go`, `port_windows.go`). Modify `ListenAndServe` to retry with port override logic. Update `main.go` callers to use the actual bound port for browser/stderr output.

**Tech Stack:** Go build tags, `lsof`/`ps` (Unix), `netstat`/`tasklist` (Windows), `net` package for port probing.

---

### Task 1: Port detection interface and process name matching

**Files:**
- Create: `internal/server/port.go`
- Test: `internal/server/port_test.go`

**Step 1: Write the failing tests**

Create `internal/server/port_test.go`:

```go
package server

import "testing"

func TestIsTgvizProcess(t *testing.T) {
	tests := []struct {
		name     string
		procName string
		want     bool
	}{
		{"exact match", "tgviz", true},
		{"with path", "/usr/local/bin/tgviz", true},
		{"windows exe", "tgviz.exe", true},
		{"full windows path", `C:\Users\me\tgviz.exe`, true},
		{"different process", "nginx", false},
		{"empty", "", false},
		{"partial false match", "tgviz-helper", true},
		{"node process", "node", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTgvizProcess(tt.procName); got != tt.want {
				t.Errorf("isTgvizProcess(%q) = %v, want %v", tt.procName, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/server/ -run TestIsTgvizProcess -v`
Expected: FAIL — `isTgvizProcess` undefined

**Step 3: Write the interface and matching function**

Create `internal/server/port.go`:

```go
package server

import "strings"

// processInfo holds details about a process using a port.
type processInfo struct {
	PID  int
	Name string
}

// portChecker finds processes on a given port and kills them.
type portChecker interface {
	findProcess(port int) (*processInfo, error)
	killProcess(pid int) error
}

// isTgvizProcess checks if a process name indicates a tgviz instance.
func isTgvizProcess(name string) bool {
	return strings.Contains(strings.ToLower(name), "tgviz")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/server/ -run TestIsTgvizProcess -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/server/port.go internal/server/port_test.go
git commit -m "feat: add port detection interface and process name matching"
```

---

### Task 2: Unix port detection implementation

**Files:**
- Create: `internal/server/port_unix.go`

**Step 1: Write the Unix implementation**

Create `internal/server/port_unix.go`:

```go
//go:build !windows

package server

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

type unixPortChecker struct{}

func newPortChecker() portChecker {
	return &unixPortChecker{}
}

func (c *unixPortChecker) findProcess(port int) (*processInfo, error) {
	out, err := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port)).Output()
	if err != nil {
		return nil, fmt.Errorf("no process found on port %d", port)
	}

	pidStr := strings.TrimSpace(strings.Split(string(out), "\n")[0])
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PID %q: %w", pidStr, err)
	}

	name := getProcessName(pid)
	return &processInfo{PID: pid, Name: name}, nil
}

func getProcessName(pid int) string {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func (c *unixPortChecker) killProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Signal(syscall.SIGTERM)
}
```

**Step 2: Verify it compiles**

Run: `go build ./internal/server/`
Expected: success

**Step 3: Commit**

```bash
git add internal/server/port_unix.go
git commit -m "feat: add Unix port detection using lsof"
```

---

### Task 3: Windows port detection implementation

**Files:**
- Create: `internal/server/port_windows.go`

**Step 1: Write the Windows implementation**

Create `internal/server/port_windows.go`:

```go
//go:build windows

package server

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type windowsPortChecker struct{}

func newPortChecker() portChecker {
	return &windowsPortChecker{}
}

var netstatPIDRe = regexp.MustCompile(`\s(\d+)\s*$`)

func (c *windowsPortChecker) findProcess(port int) (*processInfo, error) {
	out, err := exec.Command("netstat", "-ano").Output()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	needle := fmt.Sprintf(":%d ", port)
	for _, line := range strings.Split(string(out), "\n") {
		if !strings.Contains(line, needle) || !strings.Contains(line, "LISTENING") {
			continue
		}
		m := netstatPIDRe.FindStringSubmatch(line)
		if len(m) < 2 {
			continue
		}
		pid, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		name := getWindowsProcessName(pid)
		return &processInfo{PID: pid, Name: name}, nil
	}

	return nil, fmt.Errorf("no process found on port %d", port)
}

func getWindowsProcessName(pid int) string {
	out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH").Output()
	if err != nil {
		return ""
	}
	fields := strings.Split(strings.TrimSpace(string(out)), ",")
	if len(fields) > 0 {
		return strings.Trim(fields[0], "\"")
	}
	return ""
}

func (c *windowsPortChecker) killProcess(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
```

**Step 2: Verify the main package still compiles for current OS**

Run: `go build ./internal/server/`
Expected: success (Windows file excluded by build tag on macOS/Linux)

**Step 3: Commit**

```bash
git add internal/server/port_windows.go
git commit -m "feat: add Windows port detection using netstat"
```

---

### Task 4: Add retry logic to ListenAndServe

**Files:**
- Modify: `internal/server/server.go:26-36` (Server struct — add `actualPort`)
- Modify: `internal/server/server.go:60-83` (ListenAndServe — add retry loop)
- Test: `internal/server/port_test.go` (add retry logic tests)

**Step 1: Write failing tests for retry logic**

Append to `internal/server/port_test.go`:

```go
func TestResolvePort_Available(t *testing.T) {
	// Use a port that's free
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	freePort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	resolved, err := resolvePort(freePort, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved != freePort {
		t.Errorf("expected port %d, got %d", freePort, resolved)
	}
}

func TestResolvePort_Occupied_FallsBack(t *testing.T) {
	// Occupy a port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	occupiedPort := listener.Addr().(*net.TCPAddr).Port

	resolved, err := resolvePort(occupiedPort, 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved == occupiedPort {
		t.Errorf("expected a different port than %d", occupiedPort)
	}
	if resolved < occupiedPort+1 || resolved > occupiedPort+10 {
		t.Errorf("expected port in range %d-%d, got %d", occupiedPort+1, occupiedPort+10, resolved)
	}
}

func TestResolvePort_AllOccupied_Error(t *testing.T) {
	// Occupy 10 consecutive ports
	var listeners []net.Listener
	basePort := 0
	for i := 0; i < 10; i++ {
		l, err := net.Listen("tcp", fmt.Sprintf(":%d", basePort+i))
		if err != nil {
			// Can't get consecutive ports starting from 0, use dynamic
			if i == 0 {
				l2, err2 := net.Listen("tcp", ":0")
				if err2 != nil {
					t.Fatal(err2)
				}
				basePort = l2.Addr().(*net.TCPAddr).Port
				listeners = append(listeners, l2)
				continue
			}
			// Skip this test if we can't get consecutive ports
			for _, ll := range listeners {
				ll.Close()
			}
			t.Skip("cannot allocate consecutive ports")
		}
		if i == 0 {
			basePort = l.Addr().(*net.TCPAddr).Port
		}
		listeners = append(listeners, l)
	}
	defer func() {
		for _, l := range listeners {
			l.Close()
		}
	}()

	_, err := resolvePort(basePort, 10, nil)
	if err == nil {
		t.Error("expected error when all ports occupied")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/ -run TestResolvePort -v`
Expected: FAIL — `resolvePort` undefined

**Step 3: Implement resolvePort and update Server**

Add to `internal/server/port.go` (append):

```go
import (
	"fmt"
	"net"
	"strings"
	"time"
)

// resolvePort finds an available port starting from the given port.
// If the port is occupied by a tgviz process, it kills it and retries.
// If occupied by something else, it tries the next port.
// checker can be nil (uses default OS checker).
func resolvePort(startPort int, maxAttempts int, checker portChecker) (int, error) {
	if checker == nil {
		checker = newPortChecker()
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		port := startPort + attempt

		// Try to bind
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}

		// Port is busy — find out who's using it
		proc, procErr := checker.findProcess(port)
		if procErr != nil {
			// Can't determine process, skip to next port
			fmt.Fprintf(os.Stderr, "Port %d in use (unknown process), trying %d...\n", port, port+1)
			continue
		}

		if isTgvizProcess(proc.Name) {
			// Kill existing tgviz and retry same port
			fmt.Fprintf(os.Stderr, "Killing previous tgviz (PID %d) on port %d...\n", proc.PID, port)
			if killErr := checker.killProcess(proc.PID); killErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to kill PID %d: %v, trying next port...\n", proc.PID, killErr)
				continue
			}
			// Wait for port to free up
			time.Sleep(500 * time.Millisecond)
			// Retry same port
			ln2, err2 := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err2 == nil {
				ln2.Close()
				return port, nil
			}
			// Still busy, move on
			fmt.Fprintf(os.Stderr, "Port %d still in use after kill, trying %d...\n", port, port+1)
			continue
		}

		// Non-tgviz process, skip to next port
		fmt.Fprintf(os.Stderr, "Port %d in use by %s (PID %d), using %d instead\n", port, proc.Name, proc.PID, port+1)
	}

	return 0, fmt.Errorf("no available port found in range %d-%d", startPort, startPort+maxAttempts-1)
}
```

Update the imports at the top of `port.go` to include `fmt`, `net`, `os`, `strings`, `time`.

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/server/ -run TestResolvePort -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/server/port.go internal/server/port_test.go
git commit -m "feat: add port resolution with kill/fallback logic"
```

---

### Task 5: Wire resolvePort into ListenAndServe

**Files:**
- Modify: `internal/server/server.go:26-36` (add `actualPort` field)
- Modify: `internal/server/server.go:60-83` (use `resolvePort` in `ListenAndServe`)

**Step 1: Update Server struct and ListenAndServe**

In `internal/server/server.go`, add `actualPort` field to the struct:

```go
type Server struct {
	port        int
	actualPort  int
	// ... rest unchanged
}
```

Replace `ListenAndServe` with:

```go
const maxPortAttempts = 10

func (s *Server) ListenAndServe(ctx context.Context) error {
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

	select {
	case <-ctx.Done():
		return s.httpSrv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}
```

Add an accessor method:

```go
// ActualPort returns the port the server bound to (available after ListenAndServe starts).
func (s *Server) ActualPort() int {
	return s.actualPort
}
```

**Step 2: Verify existing tests still pass**

Run: `go test ./internal/server/ -v`
Expected: PASS (existing tests use httptest, not ListenAndServe directly)

**Step 3: Commit**

```bash
git add internal/server/server.go
git commit -m "feat: wire port resolution into ListenAndServe"
```

---

### Task 6: Update main.go callers to use actual port

**Files:**
- Modify: `cmd/tgviz/main.go:122-128` (planCmd — move browser/stderr after server starts)
- Modify: `cmd/tgviz/main.go:183-190` (showCmd — same)

**Step 1: Refactor planCmd server startup**

In `cmd/tgviz/main.go`, replace the plan command's server section (lines ~122-141):

```go
			// Start web UI
			srv := server.New(port, p)

			// Also output JSON for Claude Code
			if analyze {
				report := buildAnalysisReport(p)
				if err := output.PrintPlanWithReport(p, report); err != nil {
					return err
				}
			} else {
				if err := output.PrintPlan(p); err != nil {
					return err
				}
			}

			return srv.ListenAndServe(ctx, func(actualPort int) {
				fmt.Fprintf(os.Stderr, "Web UI available at http://localhost:%d\n", actualPort)
				if !noBrowser {
					server.OpenBrowser(fmt.Sprintf("http://localhost:%d", actualPort))
				}
			})
```

**Step 2: Refactor showCmd server startup**

Replace the show command's server section (lines ~183-190):

```go
			srv := server.New(port, p)
			return srv.ListenAndServe(ctx, func(actualPort int) {
				fmt.Fprintf(os.Stderr, "Web UI available at http://localhost:%d\n", actualPort)
				if !noBrowser {
					server.OpenBrowser(fmt.Sprintf("http://localhost:%d", actualPort))
				}
			})
```

**Step 3: Update ListenAndServe signature to accept onReady callback**

In `internal/server/server.go`, update `ListenAndServe`:

```go
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
```

**Step 4: Verify build succeeds**

Run: `go build ./cmd/tgviz/`
Expected: success

**Step 5: Run all tests**

Run: `go test ./... -v -count=1`
Expected: PASS

**Step 6: Commit**

```bash
git add internal/server/server.go cmd/tgviz/main.go
git commit -m "feat: use actual bound port for browser URL and stderr output"
```

---

### Task 7: Integration test and final verification

**Files:**
- Modify: `internal/server/port_test.go` (add integration-style test)

**Step 1: Add integration test — tgviz kills previous tgviz**

Append to `internal/server/port_test.go`:

```go
type mockPortChecker struct {
	process  *processInfo
	findErr  error
	killed   bool
	killErr  error
}

func (m *mockPortChecker) findProcess(port int) (*processInfo, error) {
	return m.process, m.findErr
}

func (m *mockPortChecker) killProcess(pid int) error {
	m.killed = true
	return m.killErr
}

func TestResolvePort_KillsTgviz(t *testing.T) {
	// Occupy a port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	occupiedPort := listener.Addr().(*net.TCPAddr).Port

	mock := &mockPortChecker{
		process: &processInfo{PID: 12345, Name: "tgviz"},
	}

	// resolvePort will try to bind (fail), find tgviz, kill it.
	// After kill, we close the listener to simulate the port freeing up.
	go func() {
		time.Sleep(200 * time.Millisecond)
		listener.Close()
	}()

	resolved, err := resolvePort(occupiedPort, 10, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !mock.killed {
		t.Error("expected tgviz process to be killed")
	}
	if resolved != occupiedPort {
		t.Errorf("expected same port %d after killing tgviz, got %d", occupiedPort, resolved)
	}
}

func TestResolvePort_NonTgviz_SkipsPort(t *testing.T) {
	// Occupy a port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	occupiedPort := listener.Addr().(*net.TCPAddr).Port

	mock := &mockPortChecker{
		process: &processInfo{PID: 99, Name: "nginx"},
	}

	resolved, err := resolvePort(occupiedPort, 10, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.killed {
		t.Error("should not kill non-tgviz process")
	}
	if resolved == occupiedPort {
		t.Error("should have moved to a different port")
	}
}
```

**Step 2: Run all tests**

Run: `go test ./... -v -count=1`
Expected: PASS

**Step 3: Build and verify binary**

Run: `make go && ./tgviz version`
Expected: `tgviz dev`

**Step 4: Commit**

```bash
git add internal/server/port_test.go
git commit -m "test: add mock-based tests for port resolution behavior"
```
