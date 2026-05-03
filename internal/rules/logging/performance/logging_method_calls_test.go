package loggingperformance

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingMethodCallsRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "String method call",
			code: `package test
import "log"
func test(logger Logger, addr Address) {
	logger.Info("Test", log.String("addr", addr.String()))
}`,
			wantViolations: 1,
		},
		{
			name: "Stringer field",
			code: `package test
import "log"
func test(logger Logger, addr Address) {
	logger.Info("Test", log.Stringer("addr", addr))
}`,
			wantViolations: 0,
		},
		{
			name: "String field with other call",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("addr", formatValue()))
}`,
			wantViolations: 0,
		},
		{
			name: "Nested log.String call",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("addr", log.String("nested", "value")))
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingMethodCallsRule()
			violations := rule.Check(fset, file)

			if len(violations) != tt.wantViolations {
				for i, v := range violations {
					t.Logf("Violation %d: %s at %v", i+1, v.Message, v.Position)
				}
			}
			assert.Len(t, violations, tt.wantViolations)

			for _, v := range violations {
				assert.Equal(t, rule.Name(), v.Rule)
				assert.Equal(t, violation.SeverityWarning, v.Severity)
			}
		})
	}
}

func TestLoggingMethodCallsRule_Properties(t *testing.T) {
	rule := NewLoggingMethodCallsRule()

	assert.Equal(t, "logging-method-calls", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
