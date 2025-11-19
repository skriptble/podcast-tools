package formats

import (
	"encoding/json"
	"fmt"

	"skriptble.dev/podcast-tools/models"
)

// TranscriptJSON represents the JSON structure for export
type TranscriptJSON struct {
	Segments []SegmentJSON `json:"segments"`
	Duration float64       `json:"duration"`
}

// SegmentJSON represents a single segment in JSON format
type SegmentJSON struct {
	Speaker   string  `json:"speaker"`
	Text      string  `json:"text"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

// FormatJSON formats a transcript as JSON
func FormatJSON(transcript *models.Transcript) (string, error) {
	if transcript == nil || len(transcript.Segments) == 0 {
		return "", fmt.Errorf("transcript is empty")
	}

	// Convert model segments to JSON segments
	jsonSegments := make([]SegmentJSON, len(transcript.Segments))
	for i, segment := range transcript.Segments {
		jsonSegments[i] = SegmentJSON{
			Speaker:   segment.Speaker,
			Text:      segment.Text,
			StartTime: segment.StartTime,
			EndTime:   segment.EndTime,
		}
	}

	// Create the JSON structure
	transcriptJSON := TranscriptJSON{
		Segments: jsonSegments,
		Duration: transcript.Duration(),
	}

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(transcriptJSON, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonData), nil
}
