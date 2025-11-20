package formats

import (
	"fmt"
	"strings"

	"skriptble.dev/podcast-tools/models"
)

// formatSRT formats a transcript as SRT (SubRip) subtitle format
// SRT format:
// 1
// 00:00:00,000 --> 00:00:05,000
// [Speaker]: Text
func formatSRT(transcript *models.Transcript) (string, error) {
	if transcript == nil || len(transcript.Segments) == 0 {
		return "", fmt.Errorf("transcript is empty")
	}

	var sb strings.Builder

	for i, segment := range transcript.Segments {
		// Subtitle number (1-indexed)
		sb.WriteString(fmt.Sprintf("%d\n", i+1))

		// Timestamp range (SRT uses comma for milliseconds)
		startTime := formatSRTTimestamp(segment.StartTime)
		endTime := formatSRTTimestamp(segment.EndTime)
		sb.WriteString(fmt.Sprintf("%s --> %s\n", startTime, endTime))

		// Text with speaker label
		text := strings.TrimSpace(segment.Text)
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", segment.Speaker, text))

		// Blank line between subtitles
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String()), nil
}

// formatSRTTimestamp converts seconds to SRT timestamp format (HH:MM:SS,mmm)
func formatSRTTimestamp(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}
