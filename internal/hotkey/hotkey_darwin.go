//go:build darwin

package hotkey

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon -framework ApplicationServices -framework AVFoundation -framework MediaPlayer

#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>
#import <ApplicationServices/ApplicationServices.h>
#import <AVFoundation/AVFoundation.h>
#import <MediaPlayer/MediaPlayer.h>

static id gEventMonitor = nil;
static id gKeyEventMonitor = nil;
static id gLocalKeyEventMonitor = nil;
static id gMediaEventMonitor = nil;
static id gPlayCommandTarget = nil;
static id gPauseCommandTarget = nil;
static id gToggleCommandTarget = nil;
static AVAudioPlayer *gSilentPlayer = nil;
static CFMachPortRef gSystemEventTap = NULL;
static CFRunLoopSourceRef gSystemEventTapSource = NULL;
static BOOL gHotkeyKeyDown = NO;
static BOOL gCancelKeyEnabled = NO;
static BOOL gMediaControlEnabled = NO;
static UInt16 gCurrentHotkeyCode = 0x3D;  // Default: Right Option
static UInt16 gCurrentCancelCode = 53;     // Default: Escape
static BOOL gHotkeyIsModifier = YES;       // Is the hotkey a modifier key?
static BOOL gCancelIsModifier = NO;        // Is the cancel key a modifier?

extern void goHotkeyPressed(void);
extern void goCancelPressed(void);
extern void goMediaPressed(void);

#define NX_KEYTYPE_PLAY 16
#define NX_SYSDEFINED 14

// Common key codes
#define kVK_RightOption 0x3D
#define kVK_LeftOption 0x3A
#define kVK_RightCommand 0x36
#define kVK_LeftCommand 0x37
#define kVK_RightShift 0x3C
#define kVK_LeftShift 0x38
#define kVK_RightControl 0x3E
#define kVK_LeftControl 0x3B
#define kVK_Function 0x3F
#define kVK_CapsLock 0x39
#define kVK_Escape 0x35
#define kVK_Space 0x31
#define kVK_Tab 0x30
#define kVK_Return 0x24

static void stopSystemEventTap(void);
static void stopRemoteCommandMonitoring(void);

static CGEventRef systemEventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *refcon) {
    if (!gMediaControlEnabled || type != NX_SYSDEFINED) {
        return event;
    }

    NSEvent *nsEvent = [NSEvent eventWithCGEvent:event];
    if (nsEvent == nil || [nsEvent subtype] != 8) {
        return event;
    }

    int keyCode = (([nsEvent data1] & 0xFFFF0000) >> 16);
    int keyFlags = ([nsEvent data1] & 0x0000FFFF);
    BOOL keyDown = (((keyFlags & 0xFF00) >> 8) == 0xA);
    BOOL keyRepeat = (keyFlags & 0x1);

    NSLog(@"System event tap media key: code=%d down=%d repeat=%d flags=%d", keyCode, keyDown, keyRepeat, keyFlags);

    if (keyDown && !keyRepeat && keyCode == NX_KEYTYPE_PLAY) {
        goMediaPressed();
    }

    return event;
}

// Check if accessibility permissions are granted (with optional prompt)
static int checkAccessibilityPermissionsWithPrompt(int shouldPrompt) {
    NSDictionary *options = @{(__bridge NSString *)kAXTrustedCheckOptionPrompt: @(shouldPrompt ? YES : NO)};
    BOOL trusted = AXIsProcessTrustedWithOptions((__bridge CFDictionaryRef)options);
    if (!trusted) {
        NSLog(@"Accessibility permissions not granted - please enable in System Preferences > Privacy & Security > Accessibility");
    }
    return trusted ? 1 : 0;
}

// Check if accessibility permissions are granted (no prompt)
static BOOL hasAccessibilityPermissions(void) {
    return checkAccessibilityPermissionsWithPrompt(0) ? YES : NO;
}

// Request accessibility permissions explicitly (always shows prompt)
static int requestAccessibilityPermissions(void) {
    NSLog(@"Requesting accessibility permissions...");
    int trusted = checkAccessibilityPermissionsWithPrompt(1);
    if (!trusted) {
        NSLog(@"IMPORTANT: Please grant Accessibility permissions to enable hotkey features");
        NSLog(@"Go to: System Preferences > Privacy & Security > Accessibility");
        NSLog(@"Add and enable this application");
    }
    return trusted;
}

// Request Input Monitoring permissions for lower-level media key events.
static int requestInputMonitoringPermissions(void) {
    if (@available(macOS 10.15, *)) {
        if (CGPreflightListenEventAccess()) {
            return 1;
        }
        NSLog(@"Requesting Input Monitoring permissions...");
        return CGRequestListenEventAccess() ? 1 : 0;
    }
    return 1;
}

// Check if a key code is a modifier key
static BOOL isModifierKeyCode(UInt16 keyCode) {
    switch (keyCode) {
        case kVK_RightOption:
        case kVK_LeftOption:
        case kVK_RightCommand:
        case kVK_LeftCommand:
        case kVK_RightShift:
        case kVK_LeftShift:
        case kVK_RightControl:
        case kVK_LeftControl:
        case kVK_Function:
        case kVK_CapsLock:
            return YES;
        default:
            return NO;
    }
}

// Get the modifier flag for a modifier key code
static NSEventModifierFlags getModifierFlag(UInt16 keyCode) {
    switch (keyCode) {
        case kVK_RightOption:
        case kVK_LeftOption:
            return NSEventModifierFlagOption;
        case kVK_RightCommand:
        case kVK_LeftCommand:
            return NSEventModifierFlagCommand;
        case kVK_RightShift:
        case kVK_LeftShift:
            return NSEventModifierFlagShift;
        case kVK_RightControl:
        case kVK_LeftControl:
            return NSEventModifierFlagControl;
        case kVK_Function:
            return NSEventModifierFlagFunction;
        case kVK_CapsLock:
            return NSEventModifierFlagCapsLock;
        default:
            return 0;
    }
}

static void stopAllMonitoring(void) {
    BOOL wasMediaControlEnabled = gMediaControlEnabled;
    if (gEventMonitor != nil) {
        [NSEvent removeMonitor:gEventMonitor];
        gEventMonitor = nil;
    }
    if (gKeyEventMonitor != nil) {
        [NSEvent removeMonitor:gKeyEventMonitor];
        gKeyEventMonitor = nil;
    }
    if (gLocalKeyEventMonitor != nil) {
        [NSEvent removeMonitor:gLocalKeyEventMonitor];
        gLocalKeyEventMonitor = nil;
    }
    if (!wasMediaControlEnabled && gMediaEventMonitor != nil) {
        [NSEvent removeMonitor:gMediaEventMonitor];
        gMediaEventMonitor = nil;
    }
    gHotkeyKeyDown = NO;
    gCancelKeyEnabled = NO;
}

static void stopAllMonitoringForShutdown(void) {
    gMediaControlEnabled = NO;
    stopAllMonitoring();
    stopSystemEventTap();
    stopRemoteCommandMonitoring();
}

static void startSystemEventTap(void) {
    if (gSystemEventTap != NULL) {
        return;
    }

    gSystemEventTap = CGEventTapCreate(kCGSessionEventTap,
                                       kCGHeadInsertEventTap,
                                       kCGEventTapOptionListenOnly,
                                       CGEventMaskBit(NX_SYSDEFINED),
                                       systemEventTapCallback,
                                       NULL);
    if (gSystemEventTap == NULL) {
        NSLog(@"Failed to create system event tap for media keys");
        return;
    }

    gSystemEventTapSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, gSystemEventTap, 0);
    CFRunLoopAddSource(CFRunLoopGetMain(), gSystemEventTapSource, kCFRunLoopCommonModes);
    CGEventTapEnable(gSystemEventTap, true);
    NSLog(@"System event tap started for AirPods/media control");
}

static void stopSystemEventTap(void) {
    if (gSystemEventTap != NULL) {
        CGEventTapEnable(gSystemEventTap, false);
    }
    if (gSystemEventTapSource != NULL) {
        CFRunLoopRemoveSource(CFRunLoopGetMain(), gSystemEventTapSource, kCFRunLoopCommonModes);
        CFRelease(gSystemEventTapSource);
        gSystemEventTapSource = NULL;
    }
    if (gSystemEventTap != NULL) {
        CFRelease(gSystemEventTap);
        gSystemEventTap = NULL;
    }
}

static void appendUInt16LE(NSMutableData *data, uint16_t value) {
    uint8_t bytes[2] = { value & 0xff, (value >> 8) & 0xff };
    [data appendBytes:bytes length:2];
}

static void appendUInt32LE(NSMutableData *data, uint32_t value) {
    uint8_t bytes[4] = { value & 0xff, (value >> 8) & 0xff, (value >> 16) & 0xff, (value >> 24) & 0xff };
    [data appendBytes:bytes length:4];
}

static NSData *silentWAVData(void) {
    const uint32_t sampleRate = 16000;
    const uint16_t channels = 1;
    const uint16_t bitsPerSample = 16;
    const uint32_t frames = sampleRate;
    const uint32_t dataSize = frames * channels * (bitsPerSample / 8);

    NSMutableData *data = [NSMutableData dataWithCapacity:44 + dataSize];
    [data appendBytes:"RIFF" length:4];
    appendUInt32LE(data, 36 + dataSize);
    [data appendBytes:"WAVE" length:4];
    [data appendBytes:"fmt " length:4];
    appendUInt32LE(data, 16);
    appendUInt16LE(data, 1);
    appendUInt16LE(data, channels);
    appendUInt32LE(data, sampleRate);
    appendUInt32LE(data, sampleRate * channels * (bitsPerSample / 8));
    appendUInt16LE(data, channels * (bitsPerSample / 8));
    appendUInt16LE(data, bitsPerSample);
    [data appendBytes:"data" length:4];
    appendUInt32LE(data, dataSize);
    [data increaseLengthBy:dataSize];
    return data;
}

static void startSilentPlayback(void) {
    if (gSilentPlayer != nil && gSilentPlayer.playing) {
        return;
    }

    NSError *error = nil;
    gSilentPlayer = [[AVAudioPlayer alloc] initWithData:silentWAVData() error:&error];
    if (gSilentPlayer == nil) {
        NSLog(@"Failed to create silent media player: %@", error);
        return;
    }

    gSilentPlayer.numberOfLoops = -1;
    gSilentPlayer.volume = 0.0;
    [gSilentPlayer prepareToPlay];
    [gSilentPlayer play];
    NSLog(@"Silent media playback started for AirPods control");
}

static void stopSilentPlayback(void) {
    if (gSilentPlayer != nil) {
        [gSilentPlayer stop];
        gSilentPlayer = nil;
        NSLog(@"Silent media playback stopped");
    }
}

static void stopRemoteCommandMonitoring(void) {
    MPRemoteCommandCenter *commandCenter = [MPRemoteCommandCenter sharedCommandCenter];

    if (gPlayCommandTarget != nil) {
        [commandCenter.playCommand removeTarget:gPlayCommandTarget];
        gPlayCommandTarget = nil;
    }
    if (gPauseCommandTarget != nil) {
        [commandCenter.pauseCommand removeTarget:gPauseCommandTarget];
        gPauseCommandTarget = nil;
    }
    if (gToggleCommandTarget != nil) {
        [commandCenter.togglePlayPauseCommand removeTarget:gToggleCommandTarget];
        gToggleCommandTarget = nil;
    }

    commandCenter.playCommand.enabled = NO;
    commandCenter.pauseCommand.enabled = NO;
    commandCenter.togglePlayPauseCommand.enabled = NO;
    [MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = nil;
    stopSilentPlayback();
}

static void startRemoteCommandMonitoring(void) {
    if (gToggleCommandTarget != nil) {
        startSilentPlayback();
        return;
    }

    MPRemoteCommandCenter *commandCenter = [MPRemoteCommandCenter sharedCommandCenter];
    commandCenter.playCommand.enabled = YES;
    commandCenter.pauseCommand.enabled = YES;
    commandCenter.togglePlayPauseCommand.enabled = YES;

    MPRemoteCommandHandlerStatus (^handler)(MPRemoteCommandEvent *) = ^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *event) {
        if (!gMediaControlEnabled) {
            return MPRemoteCommandHandlerStatusCommandFailed;
        }
        NSLog(@"AirPods/media remote command received: %@", event.command);
        goMediaPressed();
        startSilentPlayback();
        return MPRemoteCommandHandlerStatusSuccess;
    };

    gPlayCommandTarget = [commandCenter.playCommand addTargetWithHandler:handler];
    gPauseCommandTarget = [commandCenter.pauseCommand addTargetWithHandler:handler];
    gToggleCommandTarget = [commandCenter.togglePlayPauseCommand addTargetWithHandler:handler];

    [MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = @{
        MPMediaItemPropertyTitle: @"Yap Recording Control",
        MPMediaItemPropertyArtist: @"Yap",
        MPNowPlayingInfoPropertyPlaybackRate: @1
    };

    startSilentPlayback();
    NSLog(@"Remote command monitoring started for AirPods control");
}

static void refreshMediaControl(void) {
    if (!gMediaControlEnabled) {
        return;
    }
    startRemoteCommandMonitoring();
    startSilentPlayback();
    NSLog(@"Media control session refreshed");
}

static void startMediaMonitoring(void) {
    if (gMediaEventMonitor != nil) {
        return;
    }

    gMediaEventMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskSystemDefined
        handler:^(NSEvent *event) {
            if (!gMediaControlEnabled || [event subtype] != 8) {
                return;
            }

            int keyCode = (([event data1] & 0xFFFF0000) >> 16);
            int keyFlags = ([event data1] & 0x0000FFFF);
            BOOL keyDown = (((keyFlags & 0xFF00) >> 8) == 0xA);
            BOOL keyRepeat = (keyFlags & 0x1);

            if (keyDown && !keyRepeat && keyCode == NX_KEYTYPE_PLAY) {
                goMediaPressed();
            }
        }];

    NSLog(@"Media key monitoring started for AirPods control");
}

static void setMediaControlEnabled(int enabled) {
    gMediaControlEnabled = enabled ? YES : NO;

    if (gMediaControlEnabled) {
        startRemoteCommandMonitoring();
        startSystemEventTap();
        startMediaMonitoring();
    } else if (gMediaEventMonitor != nil) {
        [NSEvent removeMonitor:gMediaEventMonitor];
        gMediaEventMonitor = nil;
        stopRemoteCommandMonitoring();
        stopSystemEventTap();
        NSLog(@"Media key monitoring stopped");
    } else {
        stopRemoteCommandMonitoring();
        stopSystemEventTap();
    }
}

static void startMonitoring(void) {
    stopAllMonitoring();

    // Check accessibility permissions first
    if (!hasAccessibilityPermissions()) {
        NSLog(@"Cannot start monitoring without accessibility permissions");
        return;
    }

    // Monitor for modifier keys (flagsChanged events)
    gEventMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskFlagsChanged
        handler:^(NSEvent *event) {
            UInt16 keyCode = [event keyCode];
            NSEventModifierFlags flags = [event modifierFlags];

            // Check hotkey (if it's a modifier)
            if (gHotkeyIsModifier && keyCode == gCurrentHotkeyCode) {
                NSEventModifierFlags modFlag = getModifierFlag(keyCode);
                if (flags & modFlag) {
                    if (!gHotkeyKeyDown) {
                        gHotkeyKeyDown = YES;
                        goHotkeyPressed();
                    }
                } else {
                    gHotkeyKeyDown = NO;
                }
            }

            // Check cancel key (if it's a modifier)
            if (gCancelKeyEnabled && gCancelIsModifier && keyCode == gCurrentCancelCode) {
                NSEventModifierFlags modFlag = getModifierFlag(keyCode);
                if (flags & modFlag) {
                    goCancelPressed();
                }
            }
        }];

    // Monitor for regular keys (keyDown events)
    gKeyEventMonitor = [NSEvent addGlobalMonitorForEventsMatchingMask:NSEventMaskKeyDown
        handler:^(NSEvent *event) {
            UInt16 keyCode = [event keyCode];

            // Check hotkey (if it's NOT a modifier)
            if (!gHotkeyIsModifier && keyCode == gCurrentHotkeyCode) {
                goHotkeyPressed();
            }

            // Check cancel key (if it's NOT a modifier)
            if (gCancelKeyEnabled && !gCancelIsModifier && keyCode == gCurrentCancelCode) {
                goCancelPressed();
            }
        }];

    // Local monitor for when this app has focus
    gLocalKeyEventMonitor = [NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyDown
        handler:^NSEvent *(NSEvent *event) {
            UInt16 keyCode = [event keyCode];

            // Check hotkey (if it's NOT a modifier)
            if (!gHotkeyIsModifier && keyCode == gCurrentHotkeyCode) {
                goHotkeyPressed();
                return nil; // Consume event
            }

            // Check cancel key (if it's NOT a modifier)
            if (gCancelKeyEnabled && !gCancelIsModifier && keyCode == gCurrentCancelCode) {
                goCancelPressed();
                return nil; // Consume event
            }

            return event;
        }];

    NSLog(@"Key monitoring started - hotkey: %d, cancel: %d", gCurrentHotkeyCode, gCurrentCancelCode);
}

static void setHotkeyCode(UInt16 keyCode) {
    gCurrentHotkeyCode = keyCode;
    gHotkeyIsModifier = isModifierKeyCode(keyCode);
    gHotkeyKeyDown = NO;
    NSLog(@"Hotkey set to keyCode: %d (isModifier: %d)", keyCode, gHotkeyIsModifier);
}

static void setCancelCode(UInt16 keyCode) {
    gCurrentCancelCode = keyCode;
    gCancelIsModifier = isModifierKeyCode(keyCode);
    NSLog(@"Cancel key set to keyCode: %d (isModifier: %d)", keyCode, gCancelIsModifier);
}

static void enableCancelKey(void) {
    gCancelKeyEnabled = YES;
    NSLog(@"Cancel key enabled");
}

static void disableCancelKey(void) {
    gCancelKeyEnabled = NO;
    NSLog(@"Cancel key disabled");
}
*/
import "C"

import (
	"fmt"
	"strings"
	"sync"
)

var (
	callbackMu       sync.Mutex
	hotkeyCallback   func()
	hotkeyC          = make(chan struct{}, 1)
	cancelCallbackMu sync.Mutex
	cancelCallback   func()
	mediaCallbackMu  sync.Mutex
	mediaCallback    func()
)

//export goHotkeyPressed
func goHotkeyPressed() {
	select {
	case hotkeyC <- struct{}{}:
	default:
		// Channel full, dropping event
	}
}

//export goCancelPressed
func goCancelPressed() {
	cancelCallbackMu.Lock()
	cb := cancelCallback
	cancelCallbackMu.Unlock()
	if cb != nil {
		go cb()
	}
}

//export goMediaPressed
func goMediaPressed() {
	fmt.Println("AirPods/media control callback triggered")
	mediaCallbackMu.Lock()
	cb := mediaCallback
	mediaCallbackMu.Unlock()
	if cb != nil {
		go cb()
	}
}

// Callback is the function type for hotkey events
type Callback func()

// Manager handles global hotkey registration
type Manager struct {
	mu        sync.Mutex
	running   bool
	stopC     chan struct{}
	hotkeyStr string
	cancelStr string
}

// NewManager creates a new hotkey manager
func NewManager() *Manager {
	return &Manager{
		hotkeyStr: "rightoption",
		cancelStr: "escape",
	}
}

// Register registers the global hotkey
func (m *Manager) Register(cb Callback) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}

	callbackMu.Lock()
	hotkeyCallback = cb
	callbackMu.Unlock()

	// Set the hotkey code
	keyCode := KeyNameToCode(m.hotkeyStr)
	C.setHotkeyCode(C.UInt16(keyCode))

	// Set the cancel key code
	cancelCode := KeyNameToCode(m.cancelStr)
	C.setCancelCode(C.UInt16(cancelCode))

	C.startMonitoring()

	m.running = true
	m.stopC = make(chan struct{})

	// Start goroutine to handle hotkey events safely
	go func() {
		for {
			select {
			case <-hotkeyC:
				callbackMu.Lock()
				cb := hotkeyCallback
				callbackMu.Unlock()
				if cb != nil {
					cb()
				}
			case <-m.stopC:
				return
			}
		}
	}()

	return nil
}

// Unregister removes the hotkey
func (m *Manager) Unregister() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	close(m.stopC)
	C.stopAllMonitoringForShutdown()
	m.running = false

	return nil
}

// SetMediaControlEnabled starts or stops using AirPods/media play-pause as a recording control.
func (m *Manager) SetMediaControlEnabled(enabled bool, cb func()) {
	mediaCallbackMu.Lock()
	if enabled {
		mediaCallback = cb
	} else {
		mediaCallback = nil
	}
	mediaCallbackMu.Unlock()

	C.setMediaControlEnabled(C.int(boolToInt(enabled)))
}

// RefreshMediaControl re-primes the media session after recording releases the mic.
func (m *Manager) RefreshMediaControl() {
	C.refreshMediaControl()
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

// SetHotkeyType sets the recording hotkey by name
func (m *Manager) SetHotkeyType(hotkeyName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hotkeyStr = strings.ToLower(hotkeyName)
	keyCode := KeyNameToCode(m.hotkeyStr)
	C.setHotkeyCode(C.UInt16(keyCode))

	if m.running {
		C.startMonitoring()
	}

	fmt.Printf("Hotkey set to: %s (code: %d)\n", m.hotkeyStr, keyCode)
}

// SetCancelKey sets the cancel hotkey by name
func (m *Manager) SetCancelKey(keyName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cancelStr = strings.ToLower(keyName)
	cancelCode := KeyNameToCode(m.cancelStr)
	C.setCancelCode(C.UInt16(cancelCode))

	fmt.Printf("Cancel key set to: %s (code: %d)\n", m.cancelStr, cancelCode)
}

// IsRegistered returns whether hotkey is registered
func (m *Manager) IsRegistered() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// EnableCancelKey starts monitoring for the cancel key
func (m *Manager) EnableCancelKey(cb func()) {
	cancelCallbackMu.Lock()
	cancelCallback = cb
	cancelCallbackMu.Unlock()

	C.enableCancelKey()
}

// DisableCancelKey stops monitoring for the cancel key
func (m *Manager) DisableCancelKey() {
	C.disableCancelKey()

	cancelCallbackMu.Lock()
	cancelCallback = nil
	cancelCallbackMu.Unlock()
}

// GetHotkeyDisplayName returns the display name for a hotkey
func GetHotkeyDisplayName(hotkeyName string) string {
	return KeyNameToDisplayName(hotkeyName)
}

// RequestAccessibilityPermissions prompts user for accessibility permissions
func RequestAccessibilityPermissions() bool {
	return C.requestAccessibilityPermissions() != 0
}

// RequestInputMonitoringPermissions prompts user for Input Monitoring permissions.
func RequestInputMonitoringPermissions() bool {
	return C.requestInputMonitoringPermissions() != 0
}

// KeyNameToCode converts a key name to a macOS key code
func KeyNameToCode(name string) uint16 {
	switch strings.ToLower(name) {
	// Modifier keys
	case "rightoption", "rightalt":
		return 0x3D
	case "leftoption", "leftalt":
		return 0x3A
	case "rightcommand", "rightcmd":
		return 0x36
	case "leftcommand", "leftcmd":
		return 0x37
	case "rightshift":
		return 0x3C
	case "leftshift":
		return 0x38
	case "rightcontrol", "rightctrl":
		return 0x3E
	case "leftcontrol", "leftctrl":
		return 0x3B
	case "fn", "function":
		return 0x3F
	case "capslock":
		return 0x39

	// Special keys
	case "escape", "esc":
		return 0x35
	case "space":
		return 0x31
	case "tab":
		return 0x30
	case "return", "enter":
		return 0x24
	case "delete", "backspace":
		return 0x33
	case "forwarddelete":
		return 0x75

	// Arrow keys
	case "left", "arrowleft":
		return 0x7B
	case "right", "arrowright":
		return 0x7C
	case "up", "arrowup":
		return 0x7E
	case "down", "arrowdown":
		return 0x7D

	// Function keys
	case "f1":
		return 0x7A
	case "f2":
		return 0x78
	case "f3":
		return 0x63
	case "f4":
		return 0x76
	case "f5":
		return 0x60
	case "f6":
		return 0x61
	case "f7":
		return 0x62
	case "f8":
		return 0x64
	case "f9":
		return 0x65
	case "f10":
		return 0x6D
	case "f11":
		return 0x67
	case "f12":
		return 0x6F

	// Letter keys
	case "a":
		return 0x00
	case "b":
		return 0x0B
	case "c":
		return 0x08
	case "d":
		return 0x02
	case "e":
		return 0x0E
	case "f":
		return 0x03
	case "g":
		return 0x05
	case "h":
		return 0x04
	case "i":
		return 0x22
	case "j":
		return 0x26
	case "k":
		return 0x28
	case "l":
		return 0x25
	case "m":
		return 0x2E
	case "n":
		return 0x2D
	case "o":
		return 0x1F
	case "p":
		return 0x23
	case "q":
		return 0x0C
	case "r":
		return 0x0F
	case "s":
		return 0x01
	case "t":
		return 0x11
	case "u":
		return 0x20
	case "v":
		return 0x09
	case "w":
		return 0x0D
	case "x":
		return 0x07
	case "y":
		return 0x10
	case "z":
		return 0x06

	// Number keys
	case "0":
		return 0x1D
	case "1":
		return 0x12
	case "2":
		return 0x13
	case "3":
		return 0x14
	case "4":
		return 0x15
	case "5":
		return 0x17
	case "6":
		return 0x16
	case "7":
		return 0x1A
	case "8":
		return 0x1C
	case "9":
		return 0x19

	default:
		return 0x3D // Default to right option
	}
}

// KeyNameToDisplayName converts a key name to a display-friendly name
func KeyNameToDisplayName(name string) string {
	switch strings.ToLower(name) {
	case "rightoption", "rightalt":
		return "Right Option (⌥)"
	case "leftoption", "leftalt":
		return "Left Option (⌥)"
	case "rightcommand", "rightcmd":
		return "Right Command (⌘)"
	case "leftcommand", "leftcmd":
		return "Left Command (⌘)"
	case "rightshift":
		return "Right Shift (⇧)"
	case "leftshift":
		return "Left Shift (⇧)"
	case "rightcontrol", "rightctrl":
		return "Right Control (⌃)"
	case "leftcontrol", "leftctrl":
		return "Left Control (⌃)"
	case "fn", "function":
		return "Fn"
	case "capslock":
		return "Caps Lock"
	case "escape", "esc":
		return "Escape"
	case "space":
		return "Space"
	case "tab":
		return "Tab"
	case "return", "enter":
		return "Return"
	case "delete", "backspace":
		return "Delete"
	case "left", "arrowleft":
		return "←"
	case "right", "arrowright":
		return "→"
	case "up", "arrowup":
		return "↑"
	case "down", "arrowdown":
		return "↓"
	case "f1", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f9", "f10", "f11", "f12":
		return strings.ToUpper(name)
	default:
		return strings.ToUpper(name)
	}
}
