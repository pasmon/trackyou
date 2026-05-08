//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/term"
)

const detachedChildStartupGracePeriod = 500 * time.Millisecond
const detachedChildStartupPollInterval = 50 * time.Millisecond

// init runs before main() and self-detaches the process when it is launched
// directly from an interactive terminal (e.g. via `trackyou` in a shell after
// a Homebrew install).  Without this, the shell blocks until the GUI is
// closed, which feels wrong for a desktop application.
//
// How it works:
//  1. If stdin is an interactive TTY and the TRACKYOU_DETACHED marker is not
//     set, re-exec the same binary in a new session with stdin/stdout/stderr
//     disconnected and set TRACKYOU_DETACHED=1.
//  2. The parent shell sees the original process exit immediately (exit 0) and
//     returns the prompt.
//  3. The detached child process starts normally and opens the Fyne window.
//     Because the marker is already set for the child, this init() is a no-op
//     and it proceeds straight into main().
//
// Checking stdin (term.IsTerminal) rather than the controlling terminal
// (/dev/tty) is intentional: automated tools such as `brew upgrade` invoke the
// binary as a subprocess whose stdin is closed or redirected to a pipe/null,
// so IsTerminal returns false and the self-detach is correctly skipped.
func init() {
	isInteractiveTTY := false
	if os.Stdin != nil {
		isInteractiveTTY = term.IsTerminal(int(os.Stdin.Fd()))
	}

	if !shouldDetachForInteractiveLaunch(isInteractiveTTY, os.Getenv(detachMarkerEnv)) {
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Env = append(os.Environ(), detachMarkerEnv+"="+detachMarkerValue)

	if err := cmd.Start(); err != nil {
		// If re-exec fails for any reason, log and fall through to run normally.
		fmt.Fprintf(os.Stderr, "trackyou: failed to detach from terminal: %v\n", err)
		return
	}

	// Guard against silent startup failures: if the detached child exits
	// immediately, keep running in the current process so the app still opens.
	// The grace period only needs to cover early launch failures, not full app
	// readiness. 500ms is enough to catch immediate crashes while keeping the
	// parent prompt responsive.
	deadline := time.Now().Add(detachedChildStartupGracePeriod)
	for {
		var waitStatus syscall.WaitStatus
		pid, err := syscall.Wait4(cmd.Process.Pid, &waitStatus, syscall.WNOHANG, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "trackyou: detached launch status check failed (%v), retrying in foreground\n", err)
			return
		}
		if pid == cmd.Process.Pid {
			fmt.Fprintf(os.Stderr, "trackyou: detached launch exited early (status=%d), retrying in foreground\n", waitStatus)
			return
		}
		if time.Now().After(deadline) {
			os.Exit(0)
		}
		time.Sleep(detachedChildStartupPollInterval)
	}
}
