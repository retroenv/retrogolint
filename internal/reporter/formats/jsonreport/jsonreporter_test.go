package jsonreport

import (
	"bytes"
	"encoding/json"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestReporterReport_UsesStableJSONSchema(t *testing.T) {
	reporter := New()
	var output bytes.Buffer

	err := reporter.Report(&output, []violation.Violation{
		{
			Rule:     "logging-capitalization",
			Message:  "Log message should start with an uppercase letter",
			Position: token.Position{Filename: "main.go", Line: 15, Column: 2},
			Severity: violation.SeverityWarning,
		},
	})
	assert.NoError(t, err)

	var report struct {
		Violations []struct {
			Rule     string `json:"rule"`
			Position struct {
				Filename string `json:"filename"`
				Line     int    `json:"line"`
				Column   int    `json:"column"`
			} `json:"position"`
			Severity string `json:"severity"`
		} `json:"violations"`
		Count int `json:"count"`
	}
	err = json.Unmarshal(output.Bytes(), &report)
	assert.NoError(t, err)

	assert.Equal(t, 1, report.Count)
	assert.Len(t, report.Violations, 1)
	assert.Equal(t, "logging-capitalization", report.Violations[0].Rule)
	assert.Equal(t, "main.go", report.Violations[0].Position.Filename)
	assert.Equal(t, 15, report.Violations[0].Position.Line)
	assert.Equal(t, 2, report.Violations[0].Position.Column)
	assert.Equal(t, "warning", report.Violations[0].Severity)
}
