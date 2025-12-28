package transcriber

import (
	"fmt"
	"runtime"
	"sync"

	"skriptble.dev/podcast-tools/models"
)

// AudioFile represents an audio file to be transcribed
type AudioFile struct {
	Path    string // Path to the audio file
	Speaker string // Speaker label for this file
}

// ProcessConfig holds configuration for parallel processing
type ProcessConfig struct {
	AudioFiles      []AudioFile    // Audio files to process
	WhisperConfig   WhisperConfig  // Whisper configuration
	MaxParallel     int            // Maximum number of parallel transcriptions (0 = number of CPUs)
	NumTranscribers int            // Number of transcriber instances to create (0 = 1, for memory/speed tradeoff)
}

// ProcessResult holds the result of processing a single file
type ProcessResult struct {
	Speaker  string
	Segments []models.Segment
	Error    error
}

// ProcessFiles transcribes multiple audio files in parallel
func ProcessFiles(config ProcessConfig) (*models.Transcript, error) {
	if len(config.AudioFiles) == 0 {
		return nil, fmt.Errorf("no audio files provided")
	}

	// Determine number of transcriber instances
	numTranscribers := config.NumTranscribers
	if numTranscribers <= 0 {
		numTranscribers = 1
	}

	// Determine parallelism (number of workers)
	maxParallel := config.MaxParallel
	if maxParallel <= 0 {
		maxParallel = runtime.NumCPU()
	}
	// Limit workers to number of transcribers
	if maxParallel > numTranscribers {
		maxParallel = numTranscribers
	}

	if config.WhisperConfig.Verbose {
		fmt.Printf("Processing %d audio files with %d transcriber instance(s) and %d parallel worker(s)\n",
			len(config.AudioFiles), numTranscribers, maxParallel)
	}

	// Create a pool of transcriber instances
	transcriberPool := make(chan *WhisperTranscriber, numTranscribers)
	var transcribers []*WhisperTranscriber

	for i := 0; i < numTranscribers; i++ {
		transcriber, err := NewWhisperTranscriber(config.WhisperConfig)
		if err != nil {
			// Clean up any transcribers already created
			for _, t := range transcribers {
				t.Close()
			}
			return nil, fmt.Errorf("failed to create transcriber %d: %w", i+1, err)
		}
		transcribers = append(transcribers, transcriber)
		transcriberPool <- transcriber
	}
	defer func() {
		// Clean up all transcribers
		for _, t := range transcribers {
			t.Close()
		}
	}()

	// Create channels for work distribution
	jobs := make(chan AudioFile, len(config.AudioFiles))
	results := make(chan ProcessResult, len(config.AudioFiles))

	// Create worker pool
	var wg sync.WaitGroup
	for i := 0; i < maxParallel; i++ {
		wg.Add(1)
		go workerWithPool(i, transcriberPool, jobs, results, &wg)
	}

	// Send jobs to workers
	for _, audioFile := range config.AudioFiles {
		jobs <- audioFile
	}
	close(jobs)

	// Wait for all workers to finish
	wg.Wait()
	close(results)

	// Collect results
	transcript := models.NewTranscript()
	var processingErrors []error

	for result := range results {
		if result.Error != nil {
			processingErrors = append(processingErrors,
				fmt.Errorf("error processing %s: %w", result.Speaker, result.Error))
			continue
		}

		transcript.AddSegments(result.Segments)
	}

	// Check if any files were processed successfully
	if len(transcript.Segments) == 0 {
		if len(processingErrors) > 0 {
			return nil, fmt.Errorf("all transcriptions failed: %v", processingErrors)
		}
		return nil, fmt.Errorf("no segments produced from transcription")
	}

	// Sort segments by time
	transcript.SortByTime()

	if config.WhisperConfig.Verbose {
		fmt.Printf("Transcription complete: %d total segments from %d speakers\n",
			len(transcript.Segments), len(config.AudioFiles))

		if len(processingErrors) > 0 {
			fmt.Printf("Warning: %d files failed to process\n", len(processingErrors))
			for _, err := range processingErrors {
				fmt.Printf("  - %v\n", err)
			}
		}
	}

	return transcript, nil
}

// workerWithPool processes audio files from the jobs channel using transcribers from the pool
func workerWithPool(id int, transcriberPool chan *WhisperTranscriber, jobs <-chan AudioFile, results chan<- ProcessResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for audioFile := range jobs {
		// Get a transcriber from the pool
		transcriber := <-transcriberPool

		// Process the file
		segments, err := transcriber.TranscribeFile(audioFile.Path, audioFile.Speaker)

		// Return transcriber to the pool
		transcriberPool <- transcriber

		// Send result
		results <- ProcessResult{
			Speaker:  audioFile.Speaker,
			Segments: segments,
			Error:    err,
		}
	}
}

// ValidateAudioFiles checks if all audio files exist and are accessible
func ValidateAudioFiles(audioFiles []AudioFile) error {
	for i, af := range audioFiles {
		if af.Path == "" {
			return fmt.Errorf("audio file %d: path is empty", i+1)
		}
		if af.Speaker == "" {
			return fmt.Errorf("audio file %d (%s): speaker label is empty", i+1, af.Path)
		}

		// Note: We don't check file existence here as it will be checked during processing
		// This allows for more flexible error handling
	}
	return nil
}

// GenerateDefaultSpeakerLabels generates default speaker labels if not provided
func GenerateDefaultSpeakerLabels(count int) []string {
	labels := make([]string, count)
	for i := 0; i < count; i++ {
		labels[i] = fmt.Sprintf("Speaker %d", i+1)
	}
	return labels
}
