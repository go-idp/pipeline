//go:build windows

package server

import (
	"os"
	"os/exec"
)

// configureProcessGroup is a no-op on Windows.
func configureProcessGroup(cmd *exec.Cmd) {}

// killProcessGroup kills the process itself on Windows (no process groups).
func killProcessGroup(pid int) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return p.Kill()
}
