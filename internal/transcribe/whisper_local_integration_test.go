package transcribe

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLocalEngineTranscribeWAVReturnsTranscriptOnly(t *testing.T) {
	if os.Getenv("YAP_TEST_WHISPER_WAV") == "" {
		t.Skip("set YAP_TEST_WHISPER_WAV to run")
	}

	engine := NewLocalEngine(os.Getenv("YAP_TEST_WHISPER_MODELS_DIR"))
	engine.SetModel(ModelTinyEn)
	engine.SetWhisperBinary(os.Getenv("YAP_TEST_WHISPER_BIN"))

	wavData, err := os.ReadFile(os.Getenv("YAP_TEST_WHISPER_WAV"))
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	text, err := engine.TranscribeWAV(ctx, wavData)
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(text, "ggml_backend") || strings.Contains(text, "whisper_model_load") {
		t.Fatalf("transcript includes whisper diagnostics: %q", text)
	}
	if text == "" {
		t.Fatal("empty transcript")
	}
}
