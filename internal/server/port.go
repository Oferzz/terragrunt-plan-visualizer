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
