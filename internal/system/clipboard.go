package system

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/atotto/clipboard"
)

// CopyToClipboard copies text to the system clipboard
func CopyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

// ReadFromClipboard reads text from the system clipboard
func ReadFromClipboard() (string, error) {
	return clipboard.ReadAll()
}

// CopyAndPaste copies text to clipboard and simulates Cmd+V / Ctrl+V
func CopyAndPaste(text string) error {
	return CopyPasteAndSubmit(text, false)
}

// CopyPasteAndSubmit copies text, pastes it, and optionally presses Enter.
func CopyPasteAndSubmit(text string, submit bool) error {
	// First copy to clipboard
	if err := CopyToClipboard(text); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	// Then simulate paste
	if err := SimulatePaste(); err != nil {
		return err
	}

	if submit {
		return SimulateEnter()
	}
	return nil
}

// SimulatePaste simulates pressing Cmd+V (macOS) or Ctrl+V (others)
func SimulatePaste() error {
	switch runtime.GOOS {
	case "darwin":
		return simulatePasteMacOS()
	case "linux":
		return simulatePasteLinux()
	case "windows":
		return simulatePasteWindows()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// SimulateEnter simulates pressing Enter/Return.
func SimulateEnter() error {
	switch runtime.GOOS {
	case "darwin":
		return simulateEnterMacOS()
	case "linux":
		return simulateEnterLinux()
	case "windows":
		return simulateEnterWindows()
	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

// simulatePasteMacOS uses native CGEvent to simulate Cmd+V
func simulatePasteMacOS() error {
	// Use native CGEvent approach which works with the app's accessibility permissions
	return simulatePasteMacOSNative()
}

func simulateEnterMacOS() error {
	return simulateEnterMacOSNative()
}

// simulatePasteLinux uses xdotool to simulate Ctrl+V
func simulatePasteLinux() error {
	cmd := exec.Command("xdotool", "key", "ctrl+v")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to simulate paste (ensure xdotool is installed): %w", err)
	}
	return nil
}

func simulateEnterLinux() error {
	cmd := exec.Command("xdotool", "key", "Return")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to simulate enter (ensure xdotool is installed): %w", err)
	}
	return nil
}

// simulatePasteWindows uses PowerShell to simulate Ctrl+V
func simulatePasteWindows() error {
	script := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait("^v")`
	cmd := exec.Command("powershell", "-Command", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to simulate paste: %w", err)
	}
	return nil
}

func simulateEnterWindows() error {
	script := `Add-Type -AssemblyName System.Windows.Forms; [System.Windows.Forms.SendKeys]::SendWait("{ENTER}")`
	cmd := exec.Command("powershell", "-Command", script)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to simulate enter: %w", err)
	}
	return nil
}
