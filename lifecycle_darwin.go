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
	"fyne.io/fyne/v2"
)

// lifecycleRestoreFunc holds the window-restore callback set by
// registerLifecycleHandlers. It is called from Objective-C on the main thread
// whenever the screen wakes or the user session becomes active.
var lifecycleRestoreFunc func()

//export windowRestoreCallback
func windowRestoreCallback() {
	if lifecycleRestoreFunc != nil {
		fyne.Do(lifecycleRestoreFunc)
	}
}

// registerLifecycleHandlers subscribes to macOS workspace notifications so that
// the main window is automatically restored after a screen lock/unlock or a
// fast-user-switch event.
func registerLifecycleHandlers(window fyne.Window) {
	lifecycleRestoreFunc = func() {
		restoreMainWindow(window.Show, window.CenterOnScreen, window.RequestFocus)
	}
	C.TrackYouRegisterLifecycleHandlers()
}
