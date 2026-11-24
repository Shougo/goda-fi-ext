// Package detach starts the programs the kinds open (a file with xdg-open,
// a desktop file with gio, a URL in a browser) apart from Vim: Vim stops
// the process group of the plugin binary as it quits, and a program in it
// (one that does not leave it by itself) would be stopped with it.
package detach

import "os/exec"

// Start runs a program without waiting for it, in a session of its own
// (on Unix).
func Start(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	setsid(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	go func() { _ = cmd.Wait() }()
	return nil
}
