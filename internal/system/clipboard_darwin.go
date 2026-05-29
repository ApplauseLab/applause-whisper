//go:build darwin

package system

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework Carbon -framework Cocoa

#include <ApplicationServices/ApplicationServices.h>
#include <Carbon/Carbon.h>
#include <Cocoa/Cocoa.h>
#include <unistd.h>

// Check if we have accessibility permissions
int hasAccessibilityForPaste(void) {
    // Check without prompting
    BOOL trusted = AXIsProcessTrusted();
    NSLog(@"hasAccessibilityForPaste: trusted=%d", trusted);
    return trusted ? 1 : 0;
}

// Store the previous frontmost app (before our overlay appears)
static NSRunningApplication *gPreviousFrontApp = nil;

// Save the current frontmost app (call before showing overlay)
void saveFrontmostApp(void) {
    NSRunningApplication *frontApp = [[NSWorkspace sharedWorkspace] frontmostApplication];
    if (frontApp && ![frontApp.bundleIdentifier isEqualToString:@"com.wails.Yap"]) {
        gPreviousFrontApp = frontApp;
        NSLog(@"saveFrontmostApp: saved %@", frontApp.localizedName);
    }
}

// Get the frontmost application name
NSString* getFrontmostAppName(void) {
    NSRunningApplication *frontApp = [[NSWorkspace sharedWorkspace] frontmostApplication];
    return frontApp ? frontApp.localizedName : @"(unknown)";
}

// Activate the previously saved app
void activatePreviousApp(void) {
    if (gPreviousFrontApp) {
        NSLog(@"activatePreviousApp: activating %@", gPreviousFrontApp.localizedName);
        [gPreviousFrontApp activateWithOptions:NSApplicationActivateIgnoringOtherApps];
    }
}

// Simulate Cmd+V keystroke using CGEvent
int simulatePasteKeystroke(void) {
    NSLog(@"simulatePasteKeystroke: starting");
    NSLog(@"simulatePasteKeystroke: current frontmost = %@", getFrontmostAppName());

    // Check accessibility first
    if (!AXIsProcessTrusted()) {
        NSLog(@"simulatePasteKeystroke: ERROR - No accessibility permissions!");
        return 0;
    }

    // Activate the previous app first (in case our app took focus)
    if (gPreviousFrontApp) {
        NSLog(@"simulatePasteKeystroke: activating previous app %@", gPreviousFrontApp.localizedName);
        [gPreviousFrontApp activateWithOptions:NSApplicationActivateIgnoringOtherApps];
        usleep(100000); // 100ms for activation to complete
        NSLog(@"simulatePasteKeystroke: after activation, frontmost = %@", getFrontmostAppName());
    }

    // Create key down event for 'v' with Command modifier
    CGEventRef keyDown = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)kVK_ANSI_V, true);
    CGEventRef keyUp = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)kVK_ANSI_V, false);

    if (keyDown == NULL || keyUp == NULL) {
        NSLog(@"simulatePasteKeystroke: failed to create events");
        if (keyDown) CFRelease(keyDown);
        if (keyUp) CFRelease(keyUp);
        return 0;
    }

    // Set Command modifier flag for key down only. Keeping Command on key up can
    // leave the next synthetic key interpreted as a modified shortcut in some apps.
    CGEventSetFlags(keyDown, kCGEventFlagMaskCommand);
    CGEventSetFlags(keyUp, 0);

    // Post events
    NSLog(@"simulatePasteKeystroke: posting keyDown");
    CGEventPost(kCGHIDEventTap, keyDown);
    usleep(50000); // 50ms delay
    NSLog(@"simulatePasteKeystroke: posting keyUp");
    CGEventPost(kCGHIDEventTap, keyUp);

    // Release
    CFRelease(keyDown);
    CFRelease(keyUp);
    NSLog(@"simulatePasteKeystroke: done");
    return 1;
}

// Simulate Return key using CGEvent.
int simulateEnterKeystroke(void) {
    NSLog(@"simulateEnterKeystroke: starting");

    if (!AXIsProcessTrusted()) {
        NSLog(@"simulateEnterKeystroke: ERROR - No accessibility permissions!");
        return 0;
    }

    if (gPreviousFrontApp) {
        NSLog(@"simulateEnterKeystroke: activating previous app %@", gPreviousFrontApp.localizedName);
        [gPreviousFrontApp activateWithOptions:NSApplicationActivateIgnoringOtherApps];
    }

    CGEventRef keyDown = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)kVK_Return, true);
    CGEventRef keyUp = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)kVK_Return, false);

    if (keyDown == NULL || keyUp == NULL) {
        NSLog(@"simulateEnterKeystroke: failed to create events");
        if (keyDown) CFRelease(keyDown);
        if (keyUp) CFRelease(keyUp);
        return 0;
    }

    CGEventSetFlags(keyDown, 0);
    CGEventSetFlags(keyUp, 0);

    usleep(300000); // Let the paste land before submitting.
    CGEventPost(kCGHIDEventTap, keyDown);
    usleep(50000);
    CGEventPost(kCGHIDEventTap, keyUp);

    CFRelease(keyDown);
    CFRelease(keyUp);
    NSLog(@"simulateEnterKeystroke: done");
    return 1;
}
*/
import "C"

import "fmt"

// SaveFrontmostApp saves the current frontmost app (call before showing overlay)
func SaveFrontmostApp() {
	C.saveFrontmostApp()
}

// simulatePasteMacOSNative uses CGEvent to simulate Cmd+V (more reliable than AppleScript)
func simulatePasteMacOSNative() error {
	fmt.Println("simulatePasteMacOSNative: checking accessibility...")
	hasAccess := C.hasAccessibilityForPaste()
	fmt.Printf("simulatePasteMacOSNative: hasAccessibility=%d\n", hasAccess)

	if hasAccess == 0 {
		fmt.Println("ERROR: No accessibility permissions for paste! Please grant in System Settings > Privacy & Security > Accessibility")
		return fmt.Errorf("no accessibility permissions")
	}

	fmt.Println("simulatePasteMacOSNative: calling C function")
	result := C.simulatePasteKeystroke()
	fmt.Printf("simulatePasteMacOSNative: result=%d\n", result)

	if result == 0 {
		return fmt.Errorf("paste keystroke failed")
	}
	return nil
}

// simulateEnterMacOSNative uses CGEvent to simulate Return.
func simulateEnterMacOSNative() error {
	fmt.Println("simulateEnterMacOSNative: checking accessibility...")
	hasAccess := C.hasAccessibilityForPaste()
	fmt.Printf("simulateEnterMacOSNative: hasAccessibility=%d\n", hasAccess)

	if hasAccess == 0 {
		fmt.Println("ERROR: No accessibility permissions for enter! Please grant in System Settings > Privacy & Security > Accessibility")
		return fmt.Errorf("no accessibility permissions")
	}

	fmt.Println("simulateEnterMacOSNative: calling C function")
	result := C.simulateEnterKeystroke()
	fmt.Printf("simulateEnterMacOSNative: result=%d\n", result)

	if result == 0 {
		return fmt.Errorf("enter keystroke failed")
	}
	return nil
}
