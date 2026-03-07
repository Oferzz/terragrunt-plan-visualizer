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
