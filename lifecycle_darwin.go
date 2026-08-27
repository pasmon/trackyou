//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

// TrackYouLifecycleHandler is an Objective-C class that listens for screen-wake
// and user-session-active notifications from NSWorkspace and invokes a Go
// callback so the main window can be restored.
@interface TrackYouLifecycleHandler : NSObject
- (void)handleScreensDidWake:(NSNotification *)note;
- (void)handleSessionDidBecomeActive:(NSNotification *)note;
@end

// windowRestoreCallback is the Go-side function registered via
// RegisterLifecycleCallbackFunc.
static void windowRestoreCallback(void);

@implementation TrackYouLifecycleHandler
- (void)handleScreensDidWake:(NSNotification *)note {
    windowRestoreCallback();
}
- (void)handleSessionDidBecomeActive:(NSNotification *)note {
    windowRestoreCallback();
}
@end

static TrackYouLifecycleHandler *gHandler = nil;

// TrackYouRegisterLifecycleHandlers installs NSWorkspace notification observers
// for screen-wake and session-become-active events.
static void TrackYouRegisterLifecycleHandlers(void) {
    @autoreleasepool {
        if (gHandler != nil) {
            return;
        }
        gHandler = [[TrackYouLifecycleHandler alloc] init];
        NSNotificationCenter *nc = [[NSWorkspace sharedWorkspace] notificationCenter];
        [nc addObserver:gHandler
               selector:@selector(handleScreensDidWake:)
                   name:NSWorkspaceScreensDidWakeNotification
                 object:nil];
        [nc addObserver:gHandler
               selector:@selector(handleSessionDidBecomeActive:)
                   name:NSWorkspaceSessionDidBecomeActiveNotification
                 object:nil];
    }
}
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
