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
