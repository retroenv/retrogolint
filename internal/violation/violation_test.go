package violation

import (
	"encoding/json"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestShouldReport(t *testing.T) {
	tests := []struct {
		name        string
		severity    Severity
		minSeverity Severity
		want        bool
	}{
		{"equal severity", SeverityWarning, SeverityWarning, true},
		{"above minimum", SeverityError, SeverityWarning, true},
		{"below minimum", SeverityInfo, SeverityWarning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Violation{Severity: tt.severity}
			assert.Equal(t, tt.want, v.ShouldReport(tt.minSeverity))
		})
	}
}

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityError, "error"},
		{Severity(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, tt.severity.String())
	}
}

func TestSeverityMarshalJSON(t *testing.T) {
	v := Violation{
		Rule:     "test-rule",
		Message:  "test message",
		Position: token.Position{Filename: "a.go", Line: 1, Column: 1},
		Severity: SeverityError,
	}

	data, err := json.Marshal(v)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"error"`)
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  Severity
	}{
		{"info", SeverityInfo},
		{"warning", SeverityWarning},
		{"error", SeverityError},
		{"unknown", SeverityWarning},
		{"", SeverityWarning},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.want, ParseSeverity(tt.input))
	}
}
