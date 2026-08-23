//go:build !windows

package server

import (
	"os/exec"
	"syscall"
)

// configureProcessGroup starts the subprocess in its own process group so that
// cancellation can kill the whole tree (step children included).
func configureProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcessGroup kills the process group of the given process id.
func killProcessGroup(pid int) error {
	return syscall.Kill(-pid, syscall.SIGKILL)
}
