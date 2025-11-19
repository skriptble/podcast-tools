package transcriber

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"skriptble.dev/podcast-tools/models"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// WhisperConfig holds configuration for Whisper transcription
type WhisperConfig struct {
	ModelPath string  // Path to the Whisper model file
	Language  string  // Language code (e.g., "en", "es"), "auto" for detection
	Verbose   bool    // Enable verbose logging
}

// WhisperTranscriber wraps the whisper.cpp functionality
type WhisperTranscriber struct {
	model  whisper.Model
	config WhisperConfig
}

// NewWhisperTranscriber creates a new Whisper transcriber instance
func NewWhisperTranscriber(config WhisperConfig) (*WhisperTranscriber, error) {
	// Validate model path
	if config.ModelPath == "" {
		return nil, fmt.Errorf("model path is required")
	}

	// Check if model file exists
	if _, err := os.Stat(config.ModelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model file not found at %s", config.ModelPath)
	}

	// Load the Whisper model
	if config.Verbose {
		fmt.Printf("Loading Whisper model from %s...\n", config.ModelPath)
	}

	model, err := whisper.New(config.ModelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load Whisper model: %w", err)
	}

	// Set default language to auto-detect if not specified
	if config.Language == "" {
		config.Language = "auto"
	}

	if config.Verbose {
		fmt.Printf("Model loaded successfully\n")
		if model.IsMultilingual() {
			fmt.Printf("Multilingual model detected, supported languages: %d\n", len(model.Languages()))
		}
	}

	return &WhisperTranscriber{
		model:  model,
		config: config,
	}, nil
}

// Close releases resources associated with the transcriber
func (wt *WhisperTranscriber) Close() error {
	if wt.model != nil {
		return wt.model.Close()
	}
	return nil
}

// TranscribeFile transcribes an audio file and returns segments with speaker label
func (wt *WhisperTranscriber) TranscribeFile(audioPath string, speakerLabel string) ([]models.Segment, error) {
	if wt.config.Verbose {
		fmt.Printf("Transcribing %s (speaker: %s)...\n", filepath.Base(audioPath), speakerLabel)
	}

	startTime := time.Now()

	// Create a new context for this transcription
	ctx, err := wt.model.NewContext()
	if err != nil {
		return nil, fmt.Errorf("failed to create context: %w", err)
	}

	// Set language
	if wt.config.Language != "" && wt.config.Language != "auto" {
		if err := ctx.SetLanguage(wt.config.Language); err != nil {
			return nil, fmt.Errorf("failed to set language: %w", err)
		}
	}

	// Load and process the audio file
	// Note: whisper.cpp requires audio in specific format (16kHz, mono, float32)
	// We'll need to handle audio conversion in the processor
	audioData, err := loadAudioFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio file: %w", err)
	}

	// Process the audio
	if err := ctx.Process(audioData); err != nil {
		return nil, fmt.Errorf("failed to process audio: %w", err)
	}

	// Extract segments
	var segments []models.Segment
	for {
		segment, err := ctx.NextSegment()
		if err != nil {
			break // No more segments
		}

		// Convert whisper segment to our model
		seg := models.Segment{
			Speaker:   speakerLabel,
			Text:      segment.Text,
			StartTime: segment.Start.Seconds(),
			EndTime:   segment.End.Seconds(),
		}

		segments = append(segments, seg)
	}

	if wt.config.Verbose {
		duration := time.Since(startTime)
		fmt.Printf("Transcription completed for %s in %v (%d segments)\n",
			speakerLabel, duration, len(segments))
	}

	return segments, nil
}

// loadAudioFile loads an audio file and converts it to the format required by Whisper
// Whisper requires: 16kHz sample rate, mono channel, float32 PCM
func loadAudioFile(audioPath string) ([]float32, error) {
	// TODO: Implement audio file loading and conversion
	// This will need to handle various formats (WAV, MP3, M4A, FLAC)
	// and convert them to 16kHz mono float32 PCM
	// For now, we'll return an error indicating this needs to be implemented

	// Check if file exists
	if _, err := os.Stat(audioPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("audio file not found: %s", audioPath)
	}

	// For the initial implementation, we'll expect WAV files in the correct format
	// A full implementation would use a library like github.com/go-audio or ffmpeg
	return nil, fmt.Errorf("audio file loading not yet implemented - requires conversion to 16kHz mono float32 PCM")
}

// GetDefaultModelPath returns the default path for Whisper models
func GetDefaultModelPath(modelName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".cache", "whisper", fmt.Sprintf("ggml-%s.bin", modelName))
}
