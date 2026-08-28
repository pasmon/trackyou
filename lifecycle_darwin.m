#import <Cocoa/Cocoa.h>

extern void windowRestoreCallback(void);

@interface TrackYouLifecycleHandler : NSObject
- (void)handleScreensDidWake:(NSNotification *)note;
- (void)handleSessionDidBecomeActive:(NSNotification *)note;
@end

@implementation TrackYouLifecycleHandler
- (void)handleScreensDidWake:(NSNotification *)note {
    windowRestoreCallback();
}
- (void)handleSessionDidBecomeActive:(NSNotification *)note {
    windowRestoreCallback();
}
@end

static TrackYouLifecycleHandler *gHandler = nil;

void TrackYouRegisterLifecycleHandlers(void) {
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
