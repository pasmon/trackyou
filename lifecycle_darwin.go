//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

void TrackYouRegisterLifecycleHandlers(void);

// registerLifecycleHandlers stores lifecycleRestoreFunc; the Objective-C
// notifications call windowRestoreCallback, which invokes the stored Go func.
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
