//go:build linux

package detach

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

// The program is in a session of its own: not in the process group of the
// plugin binary.
func TestStartSession(t *testing.T) {
	out := filepath.Join(t.TempDir(), "sid")
	if err := Start("sh", "-c", `echo $$ > "$0.tmp" && mv "$0.tmp" "$0" && sleep 2`, out); err != nil {
		t.Fatal(err)
	}
	var b []byte
	for i := 0; i < 100; i++ {
		var err error
		if b, err = os.ReadFile(out); err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		t.Fatalf("pid: %q", b)
	}
	// A session leader: its session is its pid.
	if sid := session(t, strconv.Itoa(pid)); sid != pid {
		t.Errorf("sid: got %d, want %d (the test: %d)", sid, pid, session(t, "self"))
	}
}

// session is the session of a process: the 4th field after the name in
// /proc/<pid>/stat (state, ppid, pgrp, session).
func session(t *testing.T, pid string) int {
	t.Helper()
	b, err := os.ReadFile("/proc/" + pid + "/stat")
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	fields := strings.Fields(s[strings.LastIndexByte(s, ')')+1:])
	sid, err := strconv.Atoi(fields[3])
	if err != nil {
		t.Fatal(err)
	}
	return sid
}
