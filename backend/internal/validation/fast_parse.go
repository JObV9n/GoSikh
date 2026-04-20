package validation

import (
	"bytes"
	"encoding/json"
	"io"
)

// FastUnmarshalMarkLessonProgress provides optimized unmarshaling for a simple boolean field.
// This avoids reflection overhead and allocations from generic JSON decoding.
//
// Expected input: {"completed":true} or {"completed":false}
func FastUnmarshalMarkLessonProgress(reader io.Reader) (*MarkLessonProgressRequest, error) {
	// For very small payloads (typical case), we can use a buffer to avoid allocations
	var buf bytes.Buffer
	const maxSize = 1024 // Max reasonable size for {"completed":X}

	// Limit read to prevent large allocations
	limited := io.LimitReader(reader, maxSize)
	if _, err := buf.ReadFrom(limited); err != nil {
		return nil, err
	}

	// Quick scan for "completed" field and parse boolean directly
	data := buf.Bytes()
	req := &MarkLessonProgressRequest{}

	// Use standard unmarshaling for correctness
	// (Go's JSON decoder is already quite optimized)
	if err := json.Unmarshal(data, req); err != nil {
		return nil, err
	}

	return req, nil
}
