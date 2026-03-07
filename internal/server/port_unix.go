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
