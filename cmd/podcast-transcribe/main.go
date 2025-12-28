package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skriptble.dev/podcast-tools/formats"
	"skriptble.dev/podcast-tools/transcriber"
)

const (
	defaultModel = "large-v3"
)

var (
	// Required flags
	outputPath   = flag.String("output", "", "Output file path (required)")
	outputShort  = flag.String("o", "", "Output file path (short form)")
	formatType   = flag.String("format", "", "Output format: txt, srt, vtt, json (required)")
	formatShort  = flag.String("f", "", "Output format (short form)")

	// Optional flags
	speakers     = flag.String("speakers", "", "Comma-separated list of speaker names")
	speakersShort = flag.String("s", "", "Speaker names (short form)")
	model        = flag.String("model", defaultModel, "Whisper model: tiny, base, small, medium, large, large-v3")
	modelShort   = flag.String("m", "", "Whisper model (short form)")
	modelPath    = flag.String("model-path", "", "Path to Whisper model file (auto-detect if not provided)")
	language     = flag.String("language", "auto", "Language code (e.g., 'en', 'es') or 'auto' for detection")
	languageShort = flag.String("l", "", "Language code (short form)")
	parallel     = flag.Int("parallel", 0, "Number of parallel transcription jobs (default: number of CPU cores)")
	parallelShort = flag.Int("p", 0, "Parallel jobs (short form)")
	transcribers = flag.Int("transcribers", 0, "Number of transcriber instances for parallel processing (default: 1, each ~3GB memory)")
	transcribersShort = flag.Int("t", 0, "Transcriber instances (short form)")
	verbose      = flag.Bool("verbose", false, "Enable verbose logging")
	verboseShort = flag.Bool("v", false, "Verbose logging (short form)")
)

func main() {
	flag.Usage = printUsage
	flag.Parse()

	// Get non-flag arguments (audio files)
	audioFiles := flag.Args()

	// Validate required flags
	output := getStringFlag(*outputPath, *outputShort)
	format := getStringFlag(*formatType, *formatShort)

	if output == "" {
		fmt.Fprintln(os.Stderr, "Error: --output/-o flag is required")
		printUsage()
		os.Exit(1)
	}

	if format == "" {
		fmt.Fprintln(os.Stderr, "Error: --format/-f flag is required")
		printUsage()
		os.Exit(1)
	}

	// Validate format
	if !formats.IsValidFormat(format) {
		fmt.Fprintf(os.Stderr, "Error: invalid format '%s'. Valid formats: txt, srt, vtt, json\n", format)
		os.Exit(1)
	}

	// Validate audio files
	if len(audioFiles) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one audio file is required")
		printUsage()
		os.Exit(1)
	}

	// Get optional flags
	speakerNames := getStringFlag(*speakers, *speakersShort)
	modelName := getStringFlag(*model, *modelShort)
	if modelName == "" {
		modelName = defaultModel
	}
	lang := getStringFlag(*language, *languageShort)
	if lang == "" {
		lang = "auto"
	}
	parallelJobs := getIntFlag(*parallel, *parallelShort)
	numTranscribers := getIntFlag(*transcribers, *transcribersShort)
	isVerbose := *verbose || *verboseShort

	// Parse speaker names
	var speakerLabels []string
	if speakerNames != "" {
		speakerLabels = strings.Split(speakerNames, ",")
		// Trim whitespace from each speaker name
		for i, name := range speakerLabels {
			speakerLabels[i] = strings.TrimSpace(name)
		}
	}

	// Generate default speaker labels if not provided
	if len(speakerLabels) == 0 {
		speakerLabels = transcriber.GenerateDefaultSpeakerLabels(len(audioFiles))
	} else if len(speakerLabels) != len(audioFiles) {
		fmt.Fprintf(os.Stderr, "Error: number of speaker labels (%d) doesn't match number of audio files (%d)\n",
			len(speakerLabels), len(audioFiles))
		os.Exit(1)
	}

	// Determine model path
	modelFilePath := *modelPath
	if modelFilePath == "" {
		modelFilePath = transcriber.GetDefaultModelPath(modelName)
		if modelFilePath == "" {
			fmt.Fprintln(os.Stderr, "Error: could not determine home directory for model cache")
			os.Exit(1)
		}
	}

	// Check if model exists
	if _, err := os.Stat(modelFilePath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Whisper model not found at %s\n", modelFilePath)
		fmt.Fprintln(os.Stderr, "\nPlease download the model from:")
		fmt.Fprintf(os.Stderr, "  https://huggingface.co/ggerganov/whisper.cpp/tree/main\n")
		fmt.Fprintf(os.Stderr, "\nExpected model file: ggml-%s.bin\n", modelName)
		fmt.Fprintf(os.Stderr, "Default location: %s\n", filepath.Dir(modelFilePath))
		os.Exit(1)
	}

	if isVerbose {
		fmt.Printf("Podcast Transcription Tool\n")
		fmt.Printf("==========================\n")
		fmt.Printf("Audio files: %d\n", len(audioFiles))
		for i, file := range audioFiles {
			fmt.Printf("  %d. %s (%s)\n", i+1, file, speakerLabels[i])
		}
		fmt.Printf("Output: %s\n", output)
		fmt.Printf("Format: %s\n", format)
		fmt.Printf("Model: %s (%s)\n", modelName, modelFilePath)
		fmt.Printf("Language: %s\n", lang)
		fmt.Printf("Parallel jobs: %d\n", parallelJobs)
		fmt.Printf("Transcriber instances: %d\n", numTranscribers)
		fmt.Println()
	}

	// Prepare audio files with speaker labels
	audioFileList := make([]transcriber.AudioFile, len(audioFiles))
	for i, file := range audioFiles {
		audioFileList[i] = transcriber.AudioFile{
			Path:    file,
			Speaker: speakerLabels[i],
		}
	}

	// Validate audio files
	if err := transcriber.ValidateAudioFiles(audioFileList); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Configure processing
	config := transcriber.ProcessConfig{
		AudioFiles: audioFileList,
		WhisperConfig: transcriber.WhisperConfig{
			ModelPath: modelFilePath,
			Language:  lang,
			Verbose:   isVerbose,
		},
		MaxParallel:     parallelJobs,
		NumTranscribers: numTranscribers,
	}

	// Process files
	transcript, err := transcriber.ProcessFiles(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Format output
	formattedOutput, err := formats.FormatTranscript(transcript, formats.Format(format))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting output: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(output, []byte(formattedOutput), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output file: %v\n", err)
		os.Exit(1)
	}

	if isVerbose {
		fmt.Printf("\n✓ Transcription complete!\n")
		fmt.Printf("Output written to: %s\n", output)
		fmt.Printf("Total segments: %d\n", len(transcript.Segments))
		fmt.Printf("Total duration: %.2f seconds\n", transcript.Duration())
	} else {
		fmt.Printf("Transcription complete: %s\n", output)
	}
}

// getStringFlag returns the value from either the long or short flag (long takes precedence)
func getStringFlag(long, short string) string {
	if long != "" {
		return long
	}
	return short
}

// getIntFlag returns the value from either the long or short flag (long takes precedence)
func getIntFlag(long, short int) int {
	if long != 0 {
		return long
	}
	return short
}

// printUsage prints the usage information
func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage: podcast-transcribe [flags] <audio-file-1> <audio-file-2> [audio-file-n...]

Transcribe podcast audio files using Whisper. Each audio file should contain
a single speaker's isolated track.

Required Flags:
  --output, -o    Output file path
  --format, -f    Output format (txt, srt, vtt, json)

Optional Flags:
  --speakers, -s       Comma-separated list of speaker names (e.g., "Alice,Bob")
  --model, -m          Whisper model: tiny, base, small, medium, large, large-v3 (default: large-v3)
  --model-path         Path to Whisper model file (auto-detect if not provided)
  --language, -l       Language code (e.g., "en", "es") or "auto" for detection (default: auto)
  --parallel, -p       Number of parallel transcription jobs (default: number of CPU cores)
  --transcribers, -t   Number of transcriber instances (default: 1, each uses ~3GB memory)
  --verbose, -v        Enable verbose logging

Examples:
  # Basic usage with two speakers
  podcast-transcribe -o transcript.txt -f txt host.wav guest.wav

  # With speaker labels and SRT output
  podcast-transcribe -o transcript.srt -f srt -s "Alice,Bob" alice.wav bob.wav

  # JSON output with custom model and verbose logging
  podcast-transcribe -o transcript.json -f json -m large-v3 -v speaker1.wav speaker2.wav

  # Four speakers with 4 parallel transcriber instances for speed
  podcast-transcribe -o transcript.txt -f txt -t 4 -s "Alice,Bob,Carol,Dave" a.wav b.wav c.wav d.wav

  # Specify custom model path
  podcast-transcribe -o transcript.txt -f txt --model-path /path/to/model.bin audio.wav

Supported Formats:
  txt   Plain text with speaker labels
  srt   SubRip subtitle format with timestamps
  vtt   WebVTT subtitle format with voice tags
  json  Structured JSON with all metadata

`)
}
