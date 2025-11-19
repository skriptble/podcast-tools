# Podcast Transcription CLI Tool Specification
## Overview
Podcasts are either audio or video format, but benefit from having accurate text transcripts.
Transcription by hand takes a lot of time and effort. Thankfully, there are modern models that
perform speech recognition extremely well. Given that podcasts generally record each participant to
a separate track the task of transcription for these models is straightforward.

We need to integrate transcription into our workflow, so that we can then take transcripts and
process them through other tools to do things like automated clip finding, episode highlight
generation, and other useful tasks.

To facilitate this, we need a transcription library and CLI tool.

Build a library and command-line tool in Go that transcribes podcast audio files using Whisper. The
tool should handle multiple speaker tracks (clean, high-quality individual audio files per speaker)
and produce formatted transcripts.

## Target Hardware
- Primary: M1 Max MacBook Pro (will use Metal acceleration)
- Alternative: EPYC server with Titan X GPU (if needed)

## Technical Requirements

### Language & Dependencies
- **Language**: Go (latest stable version)
- **Speech Recognition**: whisper.cpp with Go bindings
  - Repository: https://github.com/ggerganov/whisper.cpp
  - Go bindings: Use the official Go bindings from whisper.cpp/bindings/go
  - Model: Whisper Large v3 (for best accuracy)

### Core Functionality

#### 1. Audio Input
- Accept multiple audio file paths as input (one per speaker)
- Support common audio formats: WAV, MP3, M4A, FLAC
- Each file represents a single speaker's isolated track
- Validate that files exist and are readable

#### 2. Processing
- Transcribe each speaker track independently using whisper.cpp
- Process tracks in parallel using goroutines for efficiency
- Use appropriate Whisper model (large-v3 recommended, but allow configuration)
- Extract timestamps for each transcribed segment

#### 3. Output Formats
Support multiple output formats (user selectable):
- **Plain text**: Simple merged transcript
- **SRT**: Standard subtitle format with timestamps
- **VTT**: WebVTT format
- **JSON**: Structured format with speaker labels, timestamps, and text

#### 4. Speaker Management
- Allow speaker labels/names to be provided via command-line flags
- Default to "Speaker 1", "Speaker 2", etc. if not provided
- Map each audio file to its corresponding speaker label

## CLI Interface

### Command Structure
```bash
podcast-transcribe [flags] <audio-file-1> <audio-file-2> [audio-file-n...]
```

### Required Flags
- `--output` or `-o`: Output file path
- `--format` or `-f`: Output format (txt, srt, vtt, json)

### Optional Flags
- `--speakers` or `-s`: Comma-separated list of speaker names (e.g., "Alice,Bob,Charlie")
- `--model` or `-m`: Whisper model size (tiny, base, small, medium, large, large-v3) - default: large-v3
- `--model-path`: Path to Whisper model file (auto-download if not provided)
- `--language` or `-l`: Language code (e.g., "en", "es") - auto-detect if not specified
- `--parallel` or `-p`: Number of parallel transcription jobs (default: number of CPU cores)
- `--verbose` or `-v`: Enable verbose logging

### Example Usage
```bash
# Basic usage with two speakers
podcast-transcribe -o transcript.txt -f txt host.wav guest.wav

# With speaker labels and SRT output
podcast-transcribe -o transcript.srt -f srt -s "Alice,Bob" alice.wav bob.wav

# JSON output with custom model
podcast-transcribe -o transcript.json -f json -m large-v3 --verbose speaker1.wav speaker2.wav speaker3.wav
```

## Implementation Details

### Project Structure
```
podcast-tools/
├── cmd/
│   └── podcast-transcribe/
│       └── main.go         # CLI entry point, flag parsing
├── transcriber/
│   ├── whisper.go         # Whisper integration
│   └── processor.go       # Parallel processing logic
├── formats/
│   ├── formats.go         # Format interface
│   ├── txt.go            # Plain text formatter
│   ├── srt.go            # SRT formatter
│   ├── vtt.go            # VTT formatter
│   └── json.go           # JSON formatter
├── models/
│   └── transcript.go      # Core data structures
├── Makefile               # Build automation
├── go.mod
├── go.sum
├── README.md
└── podcast-transcriber-spec.md
```

**Note**: The CLI tool is located in `cmd/podcast-transcribe/` to follow Go best practices, while the library packages (`models`, `transcriber`, `formats`) are at the top level for easy importing.

### Key Data Structures

#### Transcript Segment
```go
type Segment struct {
    Speaker   string
    Text      string
    StartTime float64  // seconds
    EndTime   float64  // seconds
}

type Transcript struct {
    Segments []Segment
}
```

### Processing Flow

1. **Parse CLI arguments** and validate inputs
2. **Initialize whisper.cpp** with the specified model
3. **Load audio files** and validate format
4. **Spawn goroutines** to transcribe each track in parallel
5. **Collect results** with speaker labels and timestamps
6. **Merge segments** from all speakers in chronological order
7. **Format output** according to selected format
8. **Write to file** or stdout

### Merging Logic
- Sort all segments by start time across all speakers
- Handle overlapping speech (if timestamps overlap, maintain chronological order by start time)
- Ensure proper formatting based on output type

### Error Handling
- Gracefully handle missing audio files
- Validate audio format compatibility
- Provide clear error messages for whisper.cpp failures
- Check for sufficient disk space for model downloads
- Handle interrupted processing with cleanup

### Performance Considerations
- Use buffered channels for segment collection
- Implement worker pool pattern for parallel processing
- Consider memory usage with large audio files
- Add progress indicators for long-running transcriptions

## Model Management
- Check if model exists locally in standard location
- Download model automatically if not present (with user confirmation)
- Use `~/.cache/whisper/` as default model directory
- Provide clear feedback during model download

## Testing Requirements
- Unit tests for formatters
- Integration test with sample audio file
- Test parallel processing with multiple files
- Validate output format correctness

## Documentation
Include a README.md with:
- Installation instructions (including whisper.cpp setup)
- Usage examples
- Supported formats and options
- Troubleshooting common issues
- Performance tips

## Build & Distribution
- Provide Makefile or build script
- Support cross-compilation for macOS (ARM64 and AMD64)
- Include instructions for building whisper.cpp dependency
- Consider packaging as Homebrew formula

## Future Enhancements (Optional)
- Real-time transcription support
- Automatic speaker diarization (for mixed audio)
- Custom vocabulary/terminology support
- Batch processing multiple podcast episodes
- Web UI for easier use

## Success Criteria
- Successfully transcribes multi-speaker podcast with high accuracy
- Produces correctly formatted output in all supported formats
- Runs efficiently on M1 Max hardware
- Clean, maintainable Go codebase
- Comprehensive error handling and user feedback
