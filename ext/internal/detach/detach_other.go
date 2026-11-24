//go:build !unix

package detach

import "os/exec"

func setsid(*exec.Cmd) {}
