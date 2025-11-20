# Claude Context: Podcast Tools

This document provides context for Claude (AI assistant) working on this project.

## Project Overview

**podcast-tools** is a collection of utilities for podcast post-production, currently featuring a multi-speaker transcription tool built with Go and whisper.cpp.

### Primary Tool: podcast-transcribe

A CLI tool and Go library for transcribing multi-speaker podcast audio using OpenAI's Whisper model via whisper.cpp bindings.

**Key Features:**
- Multi-speaker transcription with parallel processing
- Automatic audio format conversion (sample rate, channels, bit depth)
- Multiple output formats (TXT, SRT, VTT, JSON)
- Hardware acceleration support (Metal on Apple Silicon)

## Architecture

### Project Structure

```
podcast-tools/
├── cmd/podcast-transcribe/    # CLI entry point (main.go)
├── models/                     # Core data structures (Transcript, Segment)
├── transcriber/                # Whisper integration & audio processing
│   ├── whisper.go             # WAV loading, audio conversion, Whisper API
│   └── processor.go           # Parallel transcription worker pool
├── formats/                    # Output formatters
│   ├── formats.go             # Format dispatcher
│   ├── txt.go, srt.go, vtt.go, json.go  # Format implementations
├── Makefile                    # Build automation
├── go.mod                      # Module: skriptble.dev/podcast-tools
└── README.md                   # User documentation
```

### Design Decisions

1. **Library-first design**: Core functionality in top-level packages (`models`, `transcriber`, `formats`), CLI in `cmd/` subdirectory
2. **Separation of concerns**:
   - Models: Data structures only
   - Transcriber: Audio I/O and Whisper interaction
   - Formats: Output formatting (no Whisper knowledge)
   - CLI: Flag parsing and orchestration

3. **Parallel processing**: Worker pool pattern for concurrent transcription of multiple speaker tracks

## Key Implementation Details

### Audio Processing Pipeline

The audio loading happens in `transcriber/whisper.go:loadAudioFile()`:

1. **Load WAV**: Uses `github.com/go-audio/wav` decoder
2. **Stereo → Mono**: Averages channels if stereo
3. **Resample**: Linear interpolation to 16kHz (Whisper requirement)
4. **Normalize**: Converts to float32 [-1.0, 1.0] based on bit depth
5. **Pass to Whisper**: `[]float32` data to `ctx.Process()`

**Important Constants:**
- `whisper.SampleRate` = 16000 (16kHz, hardcoded in whisper.cpp)
- Sample format: float32 PCM, mono

**Supported WAV Formats:**
- 16-bit, 24-bit, or 32-bit float PCM
- Any sample rate (auto-resampled to 16kHz)
- Mono or stereo (auto-converted to mono)

### Whisper Integration

Uses official Go bindings: `github.com/ggerganov/whisper.cpp/bindings/go/pkg/whisper`

**Key API:**
- `whisper.New(modelPath)` → Model
- `model.NewContext()` → Context
- `ctx.SetLanguage(lang)` → Configure language
- `ctx.Process([]float32)` → Transcribe
- `ctx.NextSegment()` → Iterate results

**Note:** Requires CGo and compiled `libwhisper.a` (built via Makefile)

### Format Naming Convention

**Internal functions are lowercase (unexported):**
- `formatText()`, `formatSRT()`, `formatVTT()`, `formatJSON()`

**Public constants are uppercase:**
- `FormatTXT`, `FormatSRT`, `FormatVTT`, `FormatJSON`

This avoids naming conflicts between constants and functions.

### Parallel Processing

`transcriber/processor.go` implements a worker pool:
- Workers (goroutines) pull jobs from a channel
- Each worker transcribes one audio file
- Results collected via results channel
- Segments merged and sorted by timestamp

## Technical Constraints

### Whisper Requirements

1. **Sample Rate**: Must be exactly 16kHz (Whisper resamples internally)
   - We resample during WAV loading
   - Using `whisper.SampleRate` constant, not hardcoded values

2. **Format**: float32 PCM samples, mono channel
   - Normalized to [-1.0, 1.0] range

3. **Model Files**: GGML format from whisper.cpp project
   - Default location: `~/.cache/whisper/ggml-{model}.bin`
   - Download from Hugging Face: ggerganov/whisper.cpp

### Build Requirements

- **CGo enabled**: Required for whisper.cpp bindings
- **C compiler**: gcc or clang for building libwhisper.a
- **Platform**: Makefile detects platform for Metal support (macOS ARM64)

### User's Audio Specifications

The user has indicated their audio files are:
- **Format**: WAV with 32-bit float encoding
- **Channels**: Stereo
- **Sample rate**: Not specified (handled automatically)

The code now properly handles 32-bit float WAV files.

## Development Workflow

### Building

```bash
make build          # Build whisper.cpp and Go binary
make clean          # Clean build artifacts
make clean-all      # Also remove whisper.cpp directory
```

### Go Module Management

- Module name: `skriptble.dev/podcast-tools`
- Go version: 1.24 (set in go.mod)
- Toolchain: `GOTOOLCHAIN=local` to use system Go version

### Git Workflow

- Branch: `claude/transcription-tool-setup-013kvUmvA6EGXEtY2gyRRNax`
- All changes committed with detailed messages
- No main branch specified in repository

## Known Limitations & Future Work

### Current Limitations

1. **Audio Formats**: Only WAV supported (no MP3, M4A, FLAC)
   - Rationale: go-audio/wav is already a dependency; adding other formats would require additional libraries

2. **Resampling Quality**: Uses simple linear interpolation
   - Could be improved with sinc interpolation or libsamplerate

3. **No Model Download**: User must manually download Whisper models
   - Makefile has `download-model` target but requires curl

4. **Error Handling**: Continues processing other files if one fails
   - Collects errors but doesn't halt entirely

### Potential Enhancements

- Automatic speaker diarization (single mixed audio → identified speakers)
- Custom vocabulary/terminology support
- Batch processing multiple episodes
- Real-time streaming transcription
- GPU acceleration beyond Metal (CUDA, ROCm)

## Testing

Currently no tests implemented. When adding tests:
- Unit tests for formatters (easy, no external deps)
- Mock audio for transcriber tests
- Integration test would require model file (large, slow)

## Important Notes for Claude

1. **Don't modify core architecture** without user approval
   - Library/CLI separation is intentional
   - Format naming convention was fixed to avoid conflicts

2. **Audio file handling**:
   - User's files are 32-bit float stereo WAV
   - Conversion to 16kHz mono float32 is automatic
   - Don't suggest external preprocessing unless necessary

3. **Whisper.cpp dependency**:
   - Requires CGo, can't be avoided
   - Build via Makefile handles compilation
   - Can't easily test without building whisper.cpp

4. **Documentation sync**:
   - README.md is user-facing
   - podcast-transcriber-spec.md is the original specification
   - Keep both updated with implementation changes

5. **Go style**:
   - Follow standard Go conventions
   - Use `gofmt` for formatting
   - Prefer explicit over clever code

## Recent Changes

### Latest Commits

1. **Add support for 32-bit float WAV files** (9c059ef)
   - Enhanced normalization for 16/24/32-bit samples
   - Updated README to document 32-bit float support

2. **Implement WAV audio loading with automatic resampling** (6ab7bf1)
   - Complete audio pipeline from WAV to Whisper-ready float32
   - Stereo-to-mono conversion, resampling, normalization
   - Fixed naming conflicts in formats package

3. **Implement podcast transcription tool with Whisper** (ec892a3)
   - Initial implementation of library and CLI
   - Four output formats, parallel processing
   - Comprehensive Makefile and documentation

## Environment

- **Working directory**: `/home/user/podcast-tools`
- **Git repository**: Yes
- **Platform**: Linux (development), targets macOS and Linux
- **Go version**: 1.24.7 (system), 1.24 (go.mod)

## Useful Commands

```bash
# Build and verify
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go vet ./...
GOTOOLCHAIN=local gofmt -w .

# Check dependencies
go mod tidy
go mod verify

# Run CLI (after building whisper.cpp)
./build/podcast-transcribe -o test.txt -f txt -v audio.wav
```

---

*Last updated: 2025-11-20*
*Claude session: transcription-tool-setup-013kvUmvA6EGXEtY2gyRRNax*
