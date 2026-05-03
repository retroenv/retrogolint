package loggingmessage

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingCapitalizationRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "lowercase message",
			code: `package test
func test(logger Logger) {
	logger.Info("starting server")
}`,
			wantViolations: 1,
		},
		{
			name: "uppercase message",
			code: `package test
func test(logger Logger) {
	logger.Info("Starting server")
}`,
			wantViolations: 0,
		},
		{
			name: "multiple violations",
			code: `package test
func test(logger Logger) {
	logger.Info("starting server")
	logger.Debug("loading config")
	logger.Error("failed to connect")
}`,
			wantViolations: 3,
		},
		{
			name: "mixed correct and incorrect",
			code: `package test
func test(logger Logger) {
	logger.Info("Starting server")
	logger.Debug("loading config")
	logger.Info("Ready")
}`,
			wantViolations: 1,
		},
		{
			name: "empty message",
			code: `package test
func test(logger Logger) {
	logger.Info("")
}`,
			wantViolations: 0,
		},
		{
			name: "number at start",
			code: `package test
func test(logger Logger) {
	logger.Info("5 items processed")
}`,
			wantViolations: 0,
		},
		{
			name: "fmt.Errorf not flagged (false positive check)",
			code: `package test
import "fmt"
func test() error {
	return fmt.Errorf("error with lowercase: %w", nil)
}`,
			wantViolations: 0,
		},
		{
			name: "testing.T not flagged (false positive check)",
			code: `package test
func test(t *testing.T) {
	t.Errorf("test error")
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingCapitalizationRule()
			violations := rule.Check(fset, file)

			if len(violations) != tt.wantViolations {
				for i, v := range violations {
					t.Logf("Violation %d: %s at %v", i+1, v.Message, v.Position)
				}
			}
			assert.Len(t, violations, tt.wantViolations)

			// Verify violation properties
			for _, v := range violations {
				assert.Equal(t, rule.Name(), v.Rule)
				assert.Equal(t, violation.SeverityWarning, v.Severity)
			}
		})
	}
}

func TestLoggingCapitalizationRule_Properties(t *testing.T) {
	rule := NewLoggingCapitalizationRule()

	assert.Equal(t, "logging-capitalization", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
	assert.True(t, rule.Description() != "")
}
