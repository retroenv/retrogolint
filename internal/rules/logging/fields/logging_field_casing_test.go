package loggingfields

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingFieldCasingRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "camelCase field",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("fileName", "test.txt"))
}`,
			wantViolations: 1,
		},
		{
			name: "snake_case field",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("file_name", "test.txt"))
}`,
			wantViolations: 0,
		},
		{
			name: "PascalCase field",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("FileName", "test.txt"))
}`,
			wantViolations: 1,
		},
		{
			name: "multiple violations",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("fileName", "test.txt"), log.Int("userId", 123))
}`,
			wantViolations: 2,
		},
		{
			name: "mixed correct and incorrect",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("file_name", "test.txt"), log.Int("userId", 123))
}`,
			wantViolations: 1,
		},
		{
			name: "single letter field",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("x", "value"))
}`,
			wantViolations: 0,
		},
		{
			name: "field with underscores",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("file_name_test", "value"))
}`,
			wantViolations: 0,
		},
		{
			name: "field with numbers",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("test_123", "value"))
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingFieldCasingRule()
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

func TestLoggingFieldCasingRule_Properties(t *testing.T) {
	rule := NewLoggingFieldCasingRule()

	assert.Equal(t, "logging-field-casing", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
