// Package textreport holds the text reporter implementation for lint violations.
package textreport

import (
	"fmt"
	"io"

	"github.com/retroenv/retrogolint/internal/violation"
)

// Reporter formats violations as human-readable text.
type Reporter struct{}

// New constructs a text Reporter.
func New() *Reporter {
	return &Reporter{}
}

// Report writes violations in text format.
func (r *Reporter) Report(w io.Writer, violations []violation.Violation) error {
	if len(violations) == 0 {
		return nil
	}

	for _, v := range violations {
		_, err := fmt.Fprintf(w, "%s:%d:%d: %s (%s)\n",
			v.Position.Filename,
			v.Position.Line,
			v.Position.Column,
			v.Message,
			v.Rule,
		)
		if err != nil {
			return fmt.Errorf("failed to write violation: %w", err)
		}
	}

	if _, err := fmt.Fprintf(w, "\nFound %d violation(s)\n", len(violations)); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	return nil
}
