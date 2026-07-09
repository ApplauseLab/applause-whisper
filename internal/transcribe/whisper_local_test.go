package transcribe

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWhisperEnvUsesBackendFile(t *testing.T) {
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	libexecDir := filepath.Join(tmpDir, "libexec")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(libexecDir, 0755); err != nil {
		t.Fatal(err)
	}

	backendPath := filepath.Join(libexecDir, preferredAppleCPUBackend())
	if err := os.WriteFile(backendPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}
	metalPath := filepath.Join(libexecDir, "libggml-metal.so")
	if err := os.WriteFile(metalPath, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	engine := NewLocalEngine(tmpDir)
	env := engine.whisperEnv([]string{}, filepath.Join(binDir, "whisper-cli"))

	got := envValue(env, "GGML_BACKEND_PATH")
	if got != backendPath {
		t.Fatalf("GGML_BACKEND_PATH = %q, want %q", got, backendPath)
	}
}
