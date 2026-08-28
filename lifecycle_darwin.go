//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

void TrackYouRegisterLifecycleHandlers(void);

// windowRestoreCallback is the Go-side function registered via
// RegisterLifecycleCallbackFunc.
extern void windowRestoreCallback(void);
*/
import "C"

import (
	"sync"

	"fyne.io/fyne/v2"
)

var (
	lifecycleMu          sync.Mutex
	lifecycleRestoreFunc func()
)

//export windowRestoreCallback
func windowRestoreCallback() {
	lifecycleMu.Lock()
	fn := lifecycleRestoreFunc
	lifecycleMu.Unlock()
	if fn != nil {
		fyne.Do(fn)
	}
}

// registerLifecycleHandlers subscribes to macOS workspace notifications so that
// the main window is automatically restored after a screen lock/unlock or a
// fast-user-switch event.
func registerLifecycleHandlers(window fyne.Window) {
	lifecycleMu.Lock()
	lifecycleRestoreFunc = func() {
		restoreMainWindow(window.Show, window.CenterOnScreen, window.RequestFocus)
	}
	lifecycleMu.Unlock()
	C.TrackYouRegisterLifecycleHandlers()
}
