//go:build unix

package detach

import (
	"os/exec"
	"syscall"
)

func setsid(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}
