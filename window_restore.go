package main

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
