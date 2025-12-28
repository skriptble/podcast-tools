package transcriber

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-audio/wav"
	"skriptble.dev/podcast-tools/models"

	whisper "github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper"
)

// WhisperConfig holds configuration for Whisper transcription
type WhisperConfig struct {
	ModelPath string // Path to the Whisper model file
	Language  string // Language code (e.g., "en", "es"), "auto" for detection
	Verbose   bool   // Enable verbose logging
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
	// Note: whisper.cpp requires audio at whisper.SampleRate (16kHz), mono, float32
	audioData, err := loadAudioFile(audioPath, wt.config.Verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to load audio file: %w", err)
	}

	// Process the audio
	// The Process method now requires callback functions (added in newer versions of whisper.cpp)
	// We pass nil for all callbacks since we just want to iterate segments after processing
	if err := ctx.Process(audioData, nil, nil, nil); err != nil {
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

// loadAudioFile loads a WAV file and converts it to the format required by Whisper
// Whisper requires: whisper.SampleRate (16kHz), mono channel, float32 PCM
func loadAudioFile(audioPath string, verbose bool) ([]float32, error) {
	// Open the WAV file
	file, err := os.Open(audioPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	// Create WAV decoder
	decoder := wav.NewDecoder(file)
	if !decoder.IsValidFile() {
		return nil, fmt.Errorf("invalid WAV file: %s", audioPath)
	}

	// Read the audio buffer
	buf, err := decoder.FullPCMBuffer()
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	if verbose {
		fmt.Printf("  Audio format: %d Hz, %d bit, %d channel(s)\n",
			buf.Format.SampleRate, decoder.BitDepth, buf.Format.NumChannels)
	}

	// Convert to mono if stereo
	var monoSamples []int
	if buf.Format.NumChannels == 1 {
		monoSamples = buf.Data
	} else {
		// Average channels to convert to mono
		monoSamples = make([]int, len(buf.Data)/buf.Format.NumChannels)
		for i := 0; i < len(monoSamples); i++ {
			sum := 0
			for ch := 0; ch < buf.Format.NumChannels; ch++ {
				sum += buf.Data[i*buf.Format.NumChannels+ch]
			}
			monoSamples[i] = sum / buf.Format.NumChannels
		}
	}

	// Resample to whisper.SampleRate if needed
	targetRate := whisper.SampleRate
	sourceSamples := monoSamples
	if buf.Format.SampleRate != targetRate {
		if verbose {
			fmt.Printf("  Resampling from %d Hz to %d Hz\n", buf.Format.SampleRate, targetRate)
		}
		sourceSamples = resample(monoSamples, buf.Format.SampleRate, targetRate)
	}

	// Convert to float32 normalized to [-1.0, 1.0]
	floatSamples := make([]float32, len(sourceSamples))

	// Determine the normalization factor based on bit depth
	// For 16-bit: max value is 2^15 (32768)
	// For 24-bit: max value is 2^23 (8388608)
	// For 32-bit (int or float): max value is 2^31 (2147483648)
	bitDepth := decoder.BitDepth
	var maxVal float32

	switch bitDepth {
	case 16:
		maxVal = 32768.0 // 2^15
	case 24:
		maxVal = 8388608.0 // 2^23
	case 32:
		maxVal = 2147483648.0 // 2^31
	default:
		// Fallback for other bit depths
		maxVal = float32(int64(1) << uint(bitDepth-1))
	}

	for i, sample := range sourceSamples {
		floatSamples[i] = float32(sample) / maxVal

		// Clamp to [-1.0, 1.0] to handle any potential overflow
		if floatSamples[i] > 1.0 {
			floatSamples[i] = 1.0
		} else if floatSamples[i] < -1.0 {
			floatSamples[i] = -1.0
		}
	}

	if verbose {
		fmt.Printf("  Loaded %d samples (%.2f seconds)\n",
			len(floatSamples), float64(len(floatSamples))/float64(targetRate))
	}

	return floatSamples, nil
}

// resample performs simple linear interpolation resampling
func resample(samples []int, fromRate, toRate int) []int {
	if fromRate == toRate {
		return samples
	}

	ratio := float64(fromRate) / float64(toRate)
	outputLen := int(float64(len(samples)) / ratio)
	resampled := make([]int, outputLen)

	for i := 0; i < outputLen; i++ {
		srcPos := float64(i) * ratio
		srcIdx := int(srcPos)

		if srcIdx >= len(samples)-1 {
			resampled[i] = samples[len(samples)-1]
			continue
		}

		// Linear interpolation
		frac := srcPos - float64(srcIdx)
		sample1 := float64(samples[srcIdx])
		sample2 := float64(samples[srcIdx+1])
		resampled[i] = int(sample1 + (sample2-sample1)*frac)
	}

	return resampled
}

// GetDefaultModelPath returns the default path for Whisper models
func GetDefaultModelPath(modelName string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".cache", "whisper", fmt.Sprintf("ggml-%s.bin", modelName))
}
