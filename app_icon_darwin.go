//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>

#import <Cocoa/Cocoa.h>

static void TrackYouSetApplicationIcon(const void *bytes, size_t length) {
	@autoreleasepool {
		if (bytes == NULL || length <= 0) {
			return;
		}

		NSApplication *application = [NSApplication sharedApplication];
		NSData *data = [NSData dataWithBytes:bytes length:length];
		NSImage *image = [[NSImage alloc] initWithData:data];
		if (image != nil) {
			[application setApplicationIconImage:image];
			[image release];
		}
	}
}
*/
import "C"

func setPlatformApplicationIcon(iconBytes []byte) {
	if len(iconBytes) == 0 {
		return
	}

	iconData := C.CBytes(iconBytes)
	if iconData == nil {
		return
	}
	defer C.free(iconData)

	C.TrackYouSetApplicationIcon(iconData, C.size_t(len(iconBytes)))
}
