package formats

import (
	"fmt"
	"strings"

	"skriptble.dev/podcast-tools/models"
)

// FormatText formats a transcript as plain text
func FormatText(transcript *models.Transcript) (string, error) {
	if transcript == nil || len(transcript.Segments) == 0 {
		return "", fmt.Errorf("transcript is empty")
	}

	var sb strings.Builder

	currentSpeaker := ""
	for _, segment := range transcript.Segments {
		// Add speaker label when speaker changes
		if segment.Speaker != currentSpeaker {
			if currentSpeaker != "" {
				sb.WriteString("\n") // Add blank line between speakers
			}
			sb.WriteString(fmt.Sprintf("%s:\n", segment.Speaker))
			currentSpeaker = segment.Speaker
		}

		// Write the text
		sb.WriteString(strings.TrimSpace(segment.Text))
		sb.WriteString(" ")
	}

	return strings.TrimSpace(sb.String()), nil
}
