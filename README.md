# Podcast Tools

A collection of utilities to make podcast post production easier.

## Tools

### podcast-transcribe

A command-line tool for transcribing podcast audio files using OpenAI's Whisper speech recognition model via whisper.cpp. Designed for multi-speaker podcasts where each speaker is recorded on a separate audio track.

## Features

- **Multi-speaker support**: Transcribe multiple audio files, each representing a different speaker
- **Parallel processing**: Utilize multiple CPU cores for faster transcription
- **Multiple output formats**: Plain text, SRT, VTT, and JSON
- **High accuracy**: Uses Whisper Large v3 model by default
- **Hardware acceleration**: Supports Metal acceleration on Apple Silicon Macs
- **Flexible configuration**: Customizable model selection, language detection, and speaker labels

## Installation

### Prerequisites

- Go 1.24 or later
- C compiler (gcc or clang)
- Git
- Make

### Build from Source

1. Clone the repository:
```bash
git clone https://github.com/skriptble/podcast-tools.git
cd podcast-tools
```

2. Build the project (this will also build whisper.cpp):
```bash
make build
```

The binary will be created in the `build/` directory.

3. (Optional) Install to `/usr/local/bin`:
```bash
make install
```

### Download Whisper Model

Before using the tool, you need to download a Whisper model. To download the recommended large-v3 model:

```bash
make download-model
```

Alternatively, manually download from [Hugging Face](https://huggingface.co/ggerganov/whisper.cpp/tree/main) and place in `~/.cache/whisper/`.

Available models:
- `ggml-tiny.bin` - Fastest, least accurate (~75 MB)
- `ggml-base.bin` - Fast, decent accuracy (~142 MB)
- `ggml-small.bin` - Balanced (~466 MB)
- `ggml-medium.bin` - High accuracy (~1.5 GB)
- `ggml-large-v3.bin` - Best accuracy, slower (~3.1 GB) **[Recommended]**

## Usage

### Basic Usage

```bash
podcast-transcribe -o transcript.txt -f txt host.wav guest.wav
```

### With Speaker Labels

```bash
podcast-transcribe -o transcript.srt -f srt -s "Alice,Bob" alice.wav bob.wav
```

### JSON Output with Verbose Logging

```bash
podcast-transcribe -o transcript.json -f json -v speaker1.wav speaker2.wav speaker3.wav
```

### Custom Model

```bash
podcast-transcribe -o transcript.txt -f txt -m medium audio1.wav audio2.wav
```

### Specify Language

```bash
podcast-transcribe -o transcript.txt -f txt -l es spanish_speaker.wav
```

## Command-Line Options

### Required Flags

- `--output, -o` - Output file path
- `--format, -f` - Output format (txt, srt, vtt, json)

### Optional Flags

- `--speakers, -s` - Comma-separated speaker names (default: "Speaker 1", "Speaker 2", etc.)
- `--model, -m` - Whisper model size: tiny, base, small, medium, large, large-v3 (default: large-v3)
- `--model-path` - Path to Whisper model file (overrides auto-detection)
- `--language, -l` - Language code (e.g., "en", "es") or "auto" (default: auto)
- `--parallel, -p` - Number of parallel jobs (default: number of CPU cores)
- `--verbose, -v` - Enable verbose logging

## Output Formats

### Plain Text (txt)

Simple transcript with speaker labels:

```
Alice:
Hello, welcome to the show.

Bob:
Thanks for having me!
```

### SubRip (srt)

Standard subtitle format with timestamps:

```
1
00:00:00,000 --> 00:00:05,000
[Alice]: Hello, welcome to the show.

2
00:00:05,000 --> 00:00:08,500
[Bob]: Thanks for having me!
```

### WebVTT (vtt)

Web Video Text Tracks format:

```
WEBVTT

00:00:00.000 --> 00:00:05.000
<v Alice>Hello, welcome to the show.

00:00:05.000 --> 00:00:08.500
<v Bob>Thanks for having me!
```

### JSON

Structured format with all metadata:

```json
{
  "segments": [
    {
      "speaker": "Alice",
      "text": "Hello, welcome to the show.",
      "start_time": 0.0,
      "end_time": 5.0
    },
    {
      "speaker": "Bob",
      "text": "Thanks for having me!",
      "start_time": 5.0,
      "end_time": 8.5
    }
  ],
  "duration": 8.5
}
```

## Audio File Requirements

- **Format**: WAV (16-bit, 24-bit, or 32-bit float PCM)
- **Sample rate**: Any sample rate (automatically resampled to 16kHz)
- **Channels**: Mono or stereo (automatically converted to mono)
- **One file per speaker**: Each audio file should contain a single speaker's isolated track
- **Note**: Whisper internally requires 16kHz mono float32 PCM; conversion is handled automatically

## Library Usage

The transcription functionality is also available as a Go library:

```go
import (
    "skriptble.dev/podcast-tools/transcriber"
    "skriptble.dev/podcast-tools/formats"
    "skriptble.dev/podcast-tools/models"
)

// Configure transcription
config := transcriber.ProcessConfig{
    AudioFiles: []transcriber.AudioFile{
        {Path: "speaker1.wav", Speaker: "Alice"},
        {Path: "speaker2.wav", Speaker: "Bob"},
    },
    WhisperConfig: transcriber.WhisperConfig{
        ModelPath: "/path/to/ggml-large-v3.bin",
        Language:  "auto",
        Verbose:   false,
    },
    MaxParallel: 4,
}

// Process files
transcript, err := transcriber.ProcessFiles(config)
if err != nil {
    // Handle error
}

// Format output
output, err := formats.FormatTranscript(transcript, formats.FormatJSON)
```

## Performance Tips

1. **Use appropriate model size**: large-v3 for best accuracy, small/medium for faster processing
2. **Parallel processing**: The tool automatically uses all CPU cores, but you can limit with `-p`
3. **Hardware acceleration**: On M1/M2/M3 Macs, Metal acceleration is automatically enabled
4. **Audio preprocessing**: While resampling is automatic, using 16kHz mono WAV files skips conversion for slightly faster processing

## Troubleshooting

### Model not found

```
Error: Whisper model not found at ~/.cache/whisper/ggml-large-v3.bin
```

**Solution**: Run `make download-model` or manually download the model from Hugging Face.

### Build errors with whisper.cpp

```
Error: failed to load Whisper model
```

**Solution**: Ensure whisper.cpp is built correctly:
```bash
make clean-all
make build
```

### Invalid WAV file

```
Error: invalid WAV file
```

**Solution**: Ensure your audio file is a valid WAV file. If you have MP3, M4A, or other formats, convert to WAV first:
```bash
ffmpeg -i input.mp3 output.wav
```

## Project Structure

```
podcast-tools/
├── cmd/
│   └── podcast-transcribe/    # CLI entry point
│       └── main.go
├── models/                     # Core data structures
│   └── transcript.go
├── transcriber/                # Whisper integration
│   ├── whisper.go             # Whisper bindings wrapper
│   └── processor.go           # Parallel processing
├── formats/                    # Output formatters
│   ├── formats.go             # Format interface
│   ├── txt.go                 # Plain text
│   ├── srt.go                 # SubRip
│   ├── vtt.go                 # WebVTT
│   └── json.go                # JSON
├── Makefile                    # Build automation
├── go.mod                      # Go dependencies
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

[License information to be added]

## Acknowledgments

- [whisper.cpp](https://github.com/ggerganov/whisper.cpp) - High-performance Whisper implementation
- [OpenAI Whisper](https://github.com/openai/whisper) - Original speech recognition model
