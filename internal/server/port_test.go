package server

import (
	"net"
	"testing"
	"time"
)

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
		{"partial match", "tgviz-helper", true},
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

func TestResolvePort_Available(t *testing.T) {
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
}

type mockPortChecker struct {
	process *processInfo
	findErr error
	killed  bool
	killErr error
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

	// After kill, close the listener to simulate the port freeing up
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
