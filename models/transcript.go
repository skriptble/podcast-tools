package models

import (
	"fmt"
	"sort"
	"time"
)

// Segment represents a single transcribed segment with speaker information and timing
type Segment struct {
	Speaker   string    // Speaker name or label
	Text      string    // Transcribed text
	StartTime float64   // Start time in seconds
	EndTime   float64   // End time in seconds
}

// Transcript represents a complete transcript with multiple segments
type Transcript struct {
	Segments []Segment
}

// NewTranscript creates a new empty transcript
func NewTranscript() *Transcript {
	return &Transcript{
		Segments: make([]Segment, 0),
	}
}

// AddSegment adds a segment to the transcript
func (t *Transcript) AddSegment(segment Segment) {
	t.Segments = append(t.Segments, segment)
}

// AddSegments adds multiple segments to the transcript
func (t *Transcript) AddSegments(segments []Segment) {
	t.Segments = append(t.Segments, segments...)
}

// SortByTime sorts all segments chronologically by start time
func (t *Transcript) SortByTime() {
	sort.Slice(t.Segments, func(i, j int) bool {
		return t.Segments[i].StartTime < t.Segments[j].StartTime
	})
}

// Duration returns the total duration of the transcript in seconds
func (t *Transcript) Duration() float64 {
	if len(t.Segments) == 0 {
		return 0
	}

	maxEndTime := 0.0
	for _, seg := range t.Segments {
		if seg.EndTime > maxEndTime {
			maxEndTime = seg.EndTime
		}
	}
	return maxEndTime
}

// FormatTime converts seconds to a time.Duration
func FormatTime(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

// FormatTimestamp converts seconds to HH:MM:SS.mmm format
func FormatTimestamp(seconds float64) string {
	duration := FormatTime(seconds)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60
	millis := int(duration.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}
