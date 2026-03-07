package server

import (
	"net"
	"testing"
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
