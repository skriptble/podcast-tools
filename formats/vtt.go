package formats

import (
	"fmt"
	"strings"

	"skriptble.dev/podcast-tools/models"
)

// formatVTT formats a transcript as WebVTT subtitle format
// VTT format:
// WEBVTT
//
// 00:00:00.000 --> 00:00:05.000
// <v Speaker>Text
func formatVTT(transcript *models.Transcript) (string, error) {
	if transcript == nil || len(transcript.Segments) == 0 {
		return "", fmt.Errorf("transcript is empty")
	}

	var sb strings.Builder

	// VTT header
	sb.WriteString("WEBVTT\n\n")

	for _, segment := range transcript.Segments {
		// Timestamp range (VTT uses period for milliseconds)
		startTime := formatVTTTimestamp(segment.StartTime)
		endTime := formatVTTTimestamp(segment.EndTime)
		sb.WriteString(fmt.Sprintf("%s --> %s\n", startTime, endTime))

		// Text with voice tag for speaker
		text := strings.TrimSpace(segment.Text)
		sb.WriteString(fmt.Sprintf("<v %s>%s\n", segment.Speaker, text))

		// Blank line between cues
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String()), nil
}

// formatVTTTimestamp converts seconds to VTT timestamp format (HH:MM:SS.mmm)
func formatVTTTimestamp(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}
