package formats

import (
	"fmt"
	"strings"

	"skriptble.dev/podcast-tools/models"
)

// Format represents a supported output format
type Format string

const (
	FormatTXT  Format = "txt"
	FormatSRT  Format = "srt"
	FormatVTT  Format = "vtt"
	FormatJSON Format = "json"
)

// ValidFormats returns a list of all supported formats
func ValidFormats() []Format {
	return []Format{FormatTXT, FormatSRT, FormatVTT, FormatJSON}
}

// IsValidFormat checks if a format string is valid
func IsValidFormat(format string) bool {
	formatLower := Format(strings.ToLower(format))
	for _, f := range ValidFormats() {
		if f == formatLower {
			return true
		}
	}
	return false
}

// Format transcribes a transcript to the specified format
func FormatTranscript(transcript *models.Transcript, format Format) (string, error) {
	switch format {
	case FormatTXT:
		return FormatText(transcript)
	case FormatSRT:
		return FormatSRT(transcript)
	case FormatVTT:
		return FormatVTT(transcript)
	case FormatJSON:
		return FormatJSON(transcript)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
