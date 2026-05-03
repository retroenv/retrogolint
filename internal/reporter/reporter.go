// Package reporter provides formatters for outputting linting violations.
package reporter

import (
	"io"

	"github.com/retroenv/retrogolint/internal/reporter/formats/jsonreport"
	"github.com/retroenv/retrogolint/internal/reporter/formats/textreport"
	"github.com/retroenv/retrogolint/internal/violation"
)

// Reporter formats and outputs linting violations.
type Reporter interface {
	// Report writes the violations to the output writer.
	Report(w io.Writer, violations []violation.Violation) error
}

// New creates a new reporter based on the format string.
// nolint: ireturn
func New(format string) Reporter {
	switch format {
	case "json":
		return jsonreport.New()
	default:
		return textreport.New()
	}
}
