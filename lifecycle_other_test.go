//go:build !darwin

package main

import (
	"testing"

	"fyne.io/fyne/v2/test"
)

func TestRegisterLifecycleHandlers_NonDarwinNoOp(t *testing.T) {
	myApp := test.NewApp()
	window := test.NewWindow(nil)
	defer myApp.Quit()

	// registerLifecycleHandlers must not panic on non-darwin platforms.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("registerLifecycleHandlers panicked: %v", r)
		}
	}()

	registerLifecycleHandlers(window)
}
