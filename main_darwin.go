//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// init runs before main() and self-detaches the process when it is launched
// directly from a terminal (e.g. via `trackyou` in a shell after a Homebrew
// install).  Without this, the shell blocks until the GUI is closed, which
// feels wrong for a desktop application.
//
// How it works:
//  1. If stdin is connected to a controlling terminal (i.e. the process was
//     started interactively from a shell), re-exec the same binary in a new
//     session with stdin/stdout/stderr disconnected.
//  2. The parent shell sees the original process exit immediately (exit 0) and
//     returns the prompt.
//  3. The detached child process starts normally and opens the Fyne window.
//     Because the child has no controlling terminal, this init() is a no-op
//     for it and it proceeds straight into main().
func init() {
	fi, err := os.Stdin.Stat()
	if err != nil || fi.Mode()&os.ModeCharDevice == 0 {
		// Not a terminal – already detached or piped; nothing to do.
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		// If re-exec fails for any reason, log and fall through to run normally.
		fmt.Fprintf(os.Stderr, "trackyou: failed to detach from terminal: %v\n", err)
		return
	}

	os.Exit(0)
}
