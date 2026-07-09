package transcribe

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// LocalEngine uses local whisper.cpp for transcription
type LocalEngine struct {
	model      Model
	modelsDir  string
	whisperBin string // Path to whisper CLI binary
}

// NewLocalEngine creates a new local whisper.cpp engine
func NewLocalEngine(modelsDir string) *LocalEngine {
	return &LocalEngine{
		model:     ModelBaseEn,
		modelsDir: modelsDir,
	}
}

// SetWhisperBinary sets the path to the whisper CLI binary
func (e *LocalEngine) SetWhisperBinary(path string) {
	e.whisperBin = path
}

// Transcribe converts audio samples to text
func (e *LocalEngine) Transcribe(ctx context.Context, samples []float32) (string, error) {
	wavData, err := samplesToWAV(samples)
	if err != nil {
		return "", fmt.Errorf("failed to convert samples: %w", err)
	}
	return e.TranscribeWAV(ctx, wavData)
}

// TranscribeWAV transcribes WAV audio data using whisper CLI
func (e *LocalEngine) TranscribeWAV(ctx context.Context, wavData []byte) (string, error) {
	whisperBin := e.findWhisperBinary()
	if whisperBin == "" {
		return "", fmt.Errorf("whisper-cli not found. Please install whisper.cpp or set the binary path")
	}

	// Get model path
	modelPath := e.getModelPath()
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return "", fmt.Errorf("model file not found: %s. Please download the model first", modelPath)
	}

	// Create temp file for audio
	tmpFile, err := os.CreateTemp("", "whisper-input-*.wav")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(wavData); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write audio: %w", err)
	}
	tmpFile.Close()

	cmd := exec.CommandContext(ctx, whisperBin,
		"-m", modelPath,
		"-f", tmpFile.Name(),
		"--no-timestamps",
		"-otxt",
	)
	cmd.Env = e.whisperEnv(os.Environ(), whisperBin)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("whisper failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("whisper failed: %w", err)
	}

	// Clean up the output
	text := strings.TrimSpace(string(output))
	return text, nil
}

// SetModel sets the model to use
func (e *LocalEngine) SetModel(model Model) error {
	e.model = model
	return nil
}

// GetModel returns the current model
func (e *LocalEngine) GetModel() Model {
	return e.model
}

// IsAvailable checks if the engine is ready
func (e *LocalEngine) IsAvailable() bool {
	// Check if model file exists
	modelPath := e.getModelPath()
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return false
	}

	return e.findWhisperBinary() != ""
}

func (e *LocalEngine) findWhisperBinary() string {
	if e.whisperBin != "" {
		if isExecutable(e.whisperBin) {
			return e.whisperBin
		}
		if p, err := exec.LookPath(e.whisperBin); err == nil {
			return p
		}
	}

	for _, p := range bundledWhisperCandidates() {
		if isExecutable(p) {
			return p
		}
	}

	for _, p := range systemWhisperCandidates() {
		if isExecutable(p) {
			return p
		}
	}

	for _, name := range whisperBinaryNames() {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

func bundledWhisperCandidates() []string {
	exe, err := os.Executable()
	if err != nil {
		return nil
	}
	exeDir := filepath.Dir(exe)
	names := whisperBinaryNames()
	candidates := make([]string, 0, len(names)*4)
	for _, name := range names {
		switch runtime.GOOS {
		case "darwin":
			candidates = append(candidates,
				filepath.Join(exeDir, "..", "Resources", "bin", name),
				filepath.Join(exeDir, name),
			)
		case "windows":
			candidates = append(candidates,
				filepath.Join(exeDir, "bin", name),
				filepath.Join(exeDir, name),
			)
		default:
			candidates = append(candidates,
				filepath.Join(exeDir, name),
				filepath.Join(exeDir, "bin", name),
				filepath.Join(exeDir, "..", "lib", "yap", "bin", name),
			)
		}
	}
	return candidates
}

func systemWhisperCandidates() []string {
	home := os.Getenv("HOME")
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/opt/homebrew/bin/whisper-cli",
			"/usr/local/bin/whisper-cli",
			filepath.Join(home, ".local/bin/whisper-cli"),
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("ProgramFiles"), "whisper.cpp", "bin", "whisper-cli.exe"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "whisper.cpp", "whisper-cli.exe"),
		}
	default:
		return []string{
			"/usr/bin/whisper-cli",
			"/usr/local/bin/whisper-cli",
			filepath.Join(home, ".local/bin/whisper-cli"),
		}
	}
}

func whisperBinaryNames() []string {
	if runtime.GOOS == "windows" {
		return []string{"whisper-cli.exe", "whisper-cli"}
	}
	return []string{"whisper-cli"}
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func (e *LocalEngine) whisperEnv(env []string, whisperBin string) []string {
	binDir := filepath.Dir(whisperBin)
	if runtime.GOOS == "windows" {
		return append(env, "PATH="+binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	}

	backendCandidates := []string{
		filepath.Join(binDir, "..", "libexec"),
		filepath.Join(binDir, "..", "lib"),
	}
	for _, p := range backendCandidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return append(env, "GGML_BACKEND_PATH="+p)
		}
	}
	return env
}

// Name returns the provider name
func (e *LocalEngine) Name() Provider {
	return ProviderLocal
}

// getModelPath returns the full path to the model file
func (e *LocalEngine) getModelPath() string {
	modelFile := fmt.Sprintf("ggml-%s.bin", e.model)
	return filepath.Join(e.modelsDir, modelFile)
}

// GetModelPath returns the expected model file path (for download)
func (e *LocalEngine) GetModelPath() string {
	return e.getModelPath()
}

// GetModelsDir returns the models directory
func (e *LocalEngine) GetModelsDir() string {
	return e.modelsDir
}
