package main

// restoreMainWindow runs window restoration callbacks in deterministic order:
// show, then center, then focus. Each callback is optional.
func restoreMainWindow(show, center, focus func()) {
	if show != nil {
		show()
	}
	if center != nil {
		center()
	}
	if focus != nil {
		focus()
	}
}
