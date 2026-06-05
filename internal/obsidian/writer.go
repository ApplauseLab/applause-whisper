package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	yapFolder          = "Yap"
	DefaultNoteName    = "Transcriptions {{date}}"
	datePlaceholder    = "{{date}}"
	defaultDateFormat  = "2006-01-02"
	defaultMarkdownExt = ".md"
)

// AppendTranscription appends a transcript to the daily note in an Obsidian vault.
func AppendTranscription(vaultPath string, noteName string, transcript string, capturedAt time.Time) (string, error) {
	vaultPath = strings.TrimSpace(vaultPath)
	if vaultPath == "" {
		return "", fmt.Errorf("set your Obsidian vault path in Settings")
	}

	info, err := os.Stat(vaultPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("Obsidian vault folder not found")
		}
		return "", fmt.Errorf("could not read Obsidian vault folder: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("Obsidian vault path must be a folder")
	}

	date := capturedAt.Format(defaultDateFormat)
	noteDir := filepath.Join(vaultPath, yapFolder)
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		return "", fmt.Errorf("could not create Yap folder in Obsidian vault: %w", err)
	}

	fileName := ResolveNoteFileName(noteName, capturedAt)
	notePath := filepath.Join(noteDir, fileName)
	contentEntry := formatEntry(transcript, capturedAt)

	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		title := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		content := fmt.Sprintf("---\nsource: yap\ntype: transcriptions\ndate: %s\n---\n\n# %s\n\n%s", date, title, contentEntry)
		if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("could not create Obsidian note: %w", err)
		}
		return notePath, nil
	} else if err != nil {
		return "", fmt.Errorf("could not inspect Obsidian note: %w", err)
	}

	file, err := os.OpenFile(notePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("could not open Obsidian note: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(contentEntry); err != nil {
		return "", fmt.Errorf("could not append to Obsidian note: %w", err)
	}

	return notePath, nil
}

func formatEntry(transcript string, capturedAt time.Time) string {
	return fmt.Sprintf("## %s\n\n%s\n\n", capturedAt.Format("15:04"), strings.TrimSpace(transcript))
}

// ResolveNoteFileName returns a safe Markdown filename for the configured note name.
func ResolveNoteFileName(noteName string, capturedAt time.Time) string {
	name := strings.TrimSpace(noteName)
	if name == "" {
		name = DefaultNoteName
	}
	name = strings.ReplaceAll(name, datePlaceholder, capturedAt.Format(defaultDateFormat))
	name = sanitizeFileName(name)
	if !strings.HasSuffix(strings.ToLower(name), defaultMarkdownExt) {
		name += defaultMarkdownExt
	}
	return name
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	for _, invalid := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"} {
		name = strings.ReplaceAll(name, invalid, " ")
	}
	name = strings.Join(strings.Fields(name), " ")
	name = strings.Trim(name, ". ")
	if name == "" {
		return strings.ReplaceAll(DefaultNoteName, datePlaceholder, "")
	}
	return name
}

// TranscriptionsRelativePath returns the vault-relative note path for a capture date.
func TranscriptionsRelativePath(noteName string, capturedAt time.Time) string {
	return filepath.Join(yapFolder, ResolveNoteFileName(noteName, capturedAt))
}
