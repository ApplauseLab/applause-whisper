package obsidian

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const yapFolder = "Yap"

// AppendBrainCache appends a transcript to the BrainCache daily note in an Obsidian vault.
func AppendBrainCache(vaultPath string, transcript string, capturedAt time.Time) (string, error) {
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

	date := capturedAt.Format("2006-01-02")
	noteDir := filepath.Join(vaultPath, yapFolder)
	if err := os.MkdirAll(noteDir, 0755); err != nil {
		return "", fmt.Errorf("could not create Yap folder in Obsidian vault: %w", err)
	}

	notePath := filepath.Join(noteDir, fmt.Sprintf("BrainCache %s.md", date))
	contentEntry := formatEntry(transcript, capturedAt)

	if _, err := os.Stat(notePath); os.IsNotExist(err) {
		content := fmt.Sprintf("---\nsource: yap\ntype: braincache\ndate: %s\n---\n\n# BrainCache %s\n\n%s", date, date, contentEntry)
		if err := os.WriteFile(notePath, []byte(content), 0644); err != nil {
			return "", fmt.Errorf("could not create BrainCache note: %w", err)
		}
		return notePath, nil
	} else if err != nil {
		return "", fmt.Errorf("could not inspect BrainCache note: %w", err)
	}

	file, err := os.OpenFile(notePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("could not open BrainCache note: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(contentEntry); err != nil {
		return "", fmt.Errorf("could not append to BrainCache note: %w", err)
	}

	return notePath, nil
}

func formatEntry(transcript string, capturedAt time.Time) string {
	return fmt.Sprintf("## %s\n\n%s\n\n", capturedAt.Format("15:04"), strings.TrimSpace(transcript))
}

// BrainCacheRelativePath returns the vault-relative note path for a capture date.
func BrainCacheRelativePath(capturedAt time.Time) string {
	return filepath.Join(yapFolder, fmt.Sprintf("BrainCache %s.md", capturedAt.Format("2006-01-02")))
}
