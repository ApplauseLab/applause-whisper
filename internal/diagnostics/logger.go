package diagnostics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const maxLogBytes = 2 * 1024 * 1024

// Logger writes non-sensitive diagnostic events for support and troubleshooting.
type Logger struct {
	mu   sync.Mutex
	path string
}

// New creates a diagnostics logger in the app configuration directory.
func New(configDir string) *Logger {
	configDir = strings.TrimSpace(configDir)
	if configDir == "" {
		return nil
	}
	return &Logger{path: filepath.Join(configDir, "yap.log")}
}

// Info records an informational diagnostic event.
func (l *Logger) Info(event string, fields map[string]any) {
	l.write("info", event, fields)
}

// Warn records a warning diagnostic event.
func (l *Logger) Warn(event string, fields map[string]any) {
	l.write("warn", event, fields)
}

// Error records an error diagnostic event.
func (l *Logger) Error(event string, fields map[string]any) {
	l.write("error", event, fields)
}

func (l *Logger) write(level string, event string, fields map[string]any) {
	if l == nil || strings.TrimSpace(l.path) == "" {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(l.path), 0755); err != nil {
		fmt.Printf("Warning: Failed to create diagnostics directory: %v\n", err)
		return
	}
	if err := l.rotateIfNeeded(); err != nil {
		fmt.Printf("Warning: Failed to rotate diagnostics log: %v\n", err)
	}

	file, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Warning: Failed to open diagnostics log: %v\n", err)
		return
	}
	defer file.Close()

	line := fmt.Sprintf("%s level=%s event=%s", time.Now().Format(time.RFC3339), clean(level), clean(event))
	for key, value := range fields {
		line += fmt.Sprintf(" %s=%q", clean(key), cleanValue(value))
	}
	line += "\n"

	if _, err := file.WriteString(line); err != nil {
		fmt.Printf("Warning: Failed to write diagnostics log: %v\n", err)
	}
}

func (l *Logger) rotateIfNeeded() error {
	info, err := os.Stat(l.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.Size() < maxLogBytes {
		return nil
	}

	rotatedPath := l.path + ".1"
	_ = os.Remove(rotatedPath)
	return os.Rename(l.path, rotatedPath)
}

func clean(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\n", " ")
	value = strings.ReplaceAll(value, "\r", " ")
	value = strings.ReplaceAll(value, "\t", " ")
	if value == "" {
		return "unknown"
	}
	return value
}

func cleanValue(value any) string {
	text := fmt.Sprint(value)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	if len(text) > 500 {
		text = text[:500] + "..."
	}
	return text
}
