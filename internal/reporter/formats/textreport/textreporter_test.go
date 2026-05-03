package textreport

import (
	"bytes"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestReport_NoViolations(t *testing.T) {
	reporter := New()
	var buf bytes.Buffer
	err := reporter.Report(&buf, nil)
	assert.NoError(t, err)
	assert.Equal(t, "", buf.String())
}

func TestReport_SingleViolation(t *testing.T) {
	reporter := New()
	var buf bytes.Buffer

	err := reporter.Report(&buf, []violation.Violation{
		{
			Rule:     "logging-capitalization",
			Message:  "Log message should start with an uppercase letter",
			Position: token.Position{Filename: "main.go", Line: 15, Column: 2},
			Severity: violation.SeverityWarning,
		},
	})
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "main.go:15:2:")
	assert.Contains(t, out, "logging-capitalization")
	assert.Contains(t, out, "Found 1 violation(s)")
}

func TestReport_MultipleViolations(t *testing.T) {
	reporter := New()
	var buf bytes.Buffer

	violations := []violation.Violation{
		{
			Rule:     "rule-a",
			Message:  "first",
			Position: token.Position{Filename: "a.go", Line: 1, Column: 1},
		},
		{
			Rule:     "rule-b",
			Message:  "second",
			Position: token.Position{Filename: "b.go", Line: 2, Column: 3},
		},
	}
	err := reporter.Report(&buf, violations)
	assert.NoError(t, err)

	out := buf.String()
	assert.Contains(t, out, "a.go:1:1:")
	assert.Contains(t, out, "b.go:2:3:")
	assert.Contains(t, out, "Found 2 violation(s)")
}
