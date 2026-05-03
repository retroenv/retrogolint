// Package jsonreport holds the JSON reporter implementation for lint violations.
package jsonreport

import (
	"encoding/json"
	"fmt"
	"go/token"
	"io"

	"github.com/retroenv/retrogolint/internal/violation"
)

// Reporter formats violations as JSON.
type Reporter struct{}

// New constructs a JSON Reporter for violations.
func New() *Reporter {
	return &Reporter{}
}

// Report writes violations in JSON format.
func (r *Reporter) Report(w io.Writer, violations []violation.Violation) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	result := struct {
		Violations []jsonViolation `json:"violations"`
		Count      int             `json:"count"`
	}{
		Violations: toJSONViolations(violations),
		Count:      len(violations),
	}

	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

type jsonViolation struct {
	Rule     string             `json:"rule"`
	Message  string             `json:"message"`
	Position jsonPosition       `json:"position"`
	Severity violation.Severity `json:"severity"`
	Context  string             `json:"context,omitempty"`
}

type jsonPosition struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
}

func toJSONViolations(violations []violation.Violation) []jsonViolation {
	result := make([]jsonViolation, 0, len(violations))
	for _, item := range violations {
		result = append(result, jsonViolation{
			Rule:     item.Rule,
			Message:  item.Message,
			Position: toJSONPosition(item.Position),
			Severity: item.Severity,
			Context:  item.Context,
		})
	}
	return result
}

func toJSONPosition(pos token.Position) jsonPosition {
	return jsonPosition{
		Filename: pos.Filename,
		Line:     pos.Line,
		Column:   pos.Column,
	}
}
