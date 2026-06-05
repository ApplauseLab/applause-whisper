package obsidian

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendTranscriptionCreatesDailyNote(t *testing.T) {
	vaultPath := t.TempDir()
	capturedAt := time.Date(2026, 6, 4, 14, 32, 0, 0, time.Local)

	notePath, err := AppendTranscription(vaultPath, "", " Transcribed text goes here. ", capturedAt)
	if err != nil {
		t.Fatalf("AppendTranscription returned error: %v", err)
	}

	wantPath := filepath.Join(vaultPath, "Yap", "Transcriptions 2026-06-04.md")
	if notePath != wantPath {
		t.Fatalf("notePath = %q, want %q", notePath, wantPath)
	}

	contentBytes, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("failed to read note: %v", err)
	}

	want := "---\nsource: yap\ntype: transcriptions\ndate: 2026-06-04\n---\n\n# Transcriptions 2026-06-04\n\n## 14:32\n\nTranscribed text goes here.\n\n"
	if string(contentBytes) != want {
		t.Fatalf("content mismatch:\n%s", string(contentBytes))
	}
}

func TestAppendTranscriptionAppendsToExistingDailyNote(t *testing.T) {
	vaultPath := t.TempDir()
	firstCapture := time.Date(2026, 6, 4, 14, 32, 0, 0, time.Local)
	secondCapture := time.Date(2026, 6, 4, 16, 8, 0, 0, time.Local)

	if _, err := AppendTranscription(vaultPath, "", "First capture.", firstCapture); err != nil {
		t.Fatalf("first AppendTranscription returned error: %v", err)
	}
	notePath, err := AppendTranscription(vaultPath, "", "Second capture.", secondCapture)
	if err != nil {
		t.Fatalf("second AppendTranscription returned error: %v", err)
	}

	contentBytes, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("failed to read note: %v", err)
	}
	content := string(contentBytes)

	if strings.Count(content, "---\nsource: yap") != 1 {
		t.Fatalf("expected one frontmatter block, got content:\n%s", content)
	}
	if !strings.Contains(content, "## 14:32\n\nFirst capture.") {
		t.Fatalf("missing first capture:\n%s", content)
	}
	if !strings.Contains(content, "## 16:08\n\nSecond capture.") {
		t.Fatalf("missing second capture:\n%s", content)
	}

	want := "---\nsource: yap\ntype: transcriptions\ndate: 2026-06-04\n---\n\n# Transcriptions 2026-06-04\n\n## 14:32\n\nFirst capture.\n\n## 16:08\n\nSecond capture.\n\n"
	if content != want {
		t.Fatalf("content mismatch:\n%s", content)
	}
}

func TestAppendTranscriptionRequiresVaultPath(t *testing.T) {
	_, err := AppendTranscription("  ", "", "text", time.Now())
	if err == nil {
		t.Fatal("expected error for empty vault path")
	}
	if !strings.Contains(err.Error(), "set your Obsidian vault path") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppendTranscriptionRequiresVaultFolder(t *testing.T) {
	vaultFile := filepath.Join(t.TempDir(), "vault.md")
	if err := os.WriteFile(vaultFile, []byte("not a folder"), 0644); err != nil {
		t.Fatalf("failed to create vault file: %v", err)
	}

	_, err := AppendTranscription(vaultFile, "", "text", time.Now())
	if err == nil {
		t.Fatal("expected error for vault path that is not a folder")
	}
	if !strings.Contains(err.Error(), "must be a folder") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppendTranscriptionUsesCustomStaticNoteName(t *testing.T) {
	vaultPath := t.TempDir()
	capturedAt := time.Date(2026, 6, 4, 14, 32, 0, 0, time.Local)

	notePath, err := AppendTranscription(vaultPath, "Meeting Notes", "A static note.", capturedAt)
	if err != nil {
		t.Fatalf("AppendTranscription returned error: %v", err)
	}

	wantPath := filepath.Join(vaultPath, "Yap", "Meeting Notes.md")
	if notePath != wantPath {
		t.Fatalf("notePath = %q, want %q", notePath, wantPath)
	}

	contentBytes, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("failed to read note: %v", err)
	}
	if !strings.Contains(string(contentBytes), "# Meeting Notes") {
		t.Fatalf("expected custom title, got:\n%s", string(contentBytes))
	}
}

func TestResolveNoteFileNameExpandsDateAndSanitizesPath(t *testing.T) {
	capturedAt := time.Date(2026, 6, 4, 14, 32, 0, 0, time.Local)

	got := ResolveNoteFileName("Calls/{{date}}: Team", capturedAt)
	want := "Calls 2026-06-04 Team.md"
	if got != want {
		t.Fatalf("ResolveNoteFileName = %q, want %q", got, want)
	}
}
