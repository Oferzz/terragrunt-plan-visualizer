package server

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

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

const maxPortAttempts = 10

// resolvePort finds an available port starting from startPort.
// If the port is occupied by a tgviz process, it kills it and retries the same port.
// If occupied by something else, it tries the next port.
// checker can be nil (uses default OS checker).
func resolvePort(startPort int, maxAttempts int, checker portChecker) (int, error) {
	if checker == nil {
		checker = newPortChecker()
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		port := startPort + attempt

		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port, nil
		}

		proc, procErr := checker.findProcess(port)
		if procErr != nil {
			fmt.Fprintf(os.Stderr, "Port %d in use (unknown process), trying %d...\n", port, port+1)
			continue
		}

		if isTgvizProcess(proc.Name) {
			fmt.Fprintf(os.Stderr, "Killing previous tgviz (PID %d) on port %d...\n", proc.PID, port)
			if killErr := checker.killProcess(proc.PID); killErr != nil {
				fmt.Fprintf(os.Stderr, "Failed to kill PID %d: %v, trying next port...\n", proc.PID, killErr)
				continue
			}
			time.Sleep(500 * time.Millisecond)
			ln2, err2 := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err2 == nil {
				ln2.Close()
				return port, nil
			}
			fmt.Fprintf(os.Stderr, "Port %d still in use after kill, trying %d...\n", port, port+1)
			continue
		}

		fmt.Fprintf(os.Stderr, "Port %d in use by %s (PID %d), using %d instead\n", port, proc.Name, proc.PID, port+1)
	}

	return 0, fmt.Errorf("no available port found in range %d-%d", startPort, startPort+maxAttempts-1)
}
