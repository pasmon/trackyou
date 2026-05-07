//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

#import <Cocoa/Cocoa.h>

static void trackyouSetApplicationIcon(const void *bytes, int length) {
	@autoreleasepool {
		if (bytes == NULL || length <= 0) {
			return;
		}

		NSApplication *application = [NSApplication sharedApplication];
		NSData *data = [NSData dataWithBytes:bytes length:(NSUInteger) length];
		NSImage *image = [[NSImage alloc] initWithData:data];
		if (image != nil) {
			[application setApplicationIconImage:image];
		}
	}
}
*/
import "C"

import "unsafe"

func setPlatformApplicationIcon(iconBytes []byte) {
	if len(iconBytes) == 0 {
		return
	}

	C.trackyouSetApplicationIcon(unsafe.Pointer(&iconBytes[0]), C.int(len(iconBytes)))
}
