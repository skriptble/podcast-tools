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

// FormatTranscript transcribes a transcript to the specified format
func FormatTranscript(transcript *models.Transcript, format Format) (string, error) {
	switch format {
	case FormatTXT:
		return formatText(transcript)
	case FormatSRT:
		return formatSRT(transcript)
	case FormatVTT:
		return formatVTT(transcript)
	case FormatJSON:
		return formatJSON(transcript)
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
