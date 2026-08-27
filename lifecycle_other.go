//go:build !darwin

package main

import "fyne.io/fyne/v2"

// registerLifecycleHandlers is a no-op on non-macOS platforms.
func registerLifecycleHandlers(_ fyne.Window) {}
