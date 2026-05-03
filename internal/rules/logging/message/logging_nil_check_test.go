package loggingmessage

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingNilCheckRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "unnecessary nil check",
			code: `package test
func test(logger *Logger) {
	if logger != nil {
		logger.Info("Test")
	}
}`,
			wantViolations: 1,
		},
		{
			name: "unnecessary nil check (reversed)",
			code: `package test
func test(logger *Logger) {
	if nil != logger {
		logger.Info("Test")
	}
}`,
			wantViolations: 1,
		},
		{
			name: "no nil check",
			code: `package test
func test(logger *Logger) {
	logger.Info("Test")
}`,
			wantViolations: 0,
		},
		{
			name: "nil check with multiple logger calls",
			code: `package test
func test(logger *Logger) {
	if logger != nil {
		logger.Info("Starting")
		logger.Debug("Processing")
	}
}`,
			wantViolations: 1,
		},
		{
			name: "nil check with other statements (not a violation)",
			code: `package test
func test(logger *Logger) {
	if logger != nil {
		logger.Info("Test")
		x := 5
		_ = x
	}
}`,
			wantViolations: 0,
		},
		{
			name: "selector nil check",
			code: `package test
type Mapper struct {
	logger *Logger
}

func test(mapper *Mapper) {
	if mapper.logger != nil {
		mapper.logger.Info("Processing")
	}
}`,
			wantViolations: 1,
		},
		{
			name: "different variable check (not a violation)",
			code: `package test
func test(logger *Logger, other *Logger) {
	if other != nil {
		logger.Info("Test")
	}
}`,
			wantViolations: 0,
		},
		{
			name: "multiple unnecessary nil checks",
			code: `package test
func test(logger *Logger) {
	if logger != nil {
		logger.Info("Test")
	}
	if logger != nil {
		logger.Debug("Debug")
	}
}`,
			wantViolations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingNilCheckRule()
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

func TestLoggingNilCheckRule_Properties(t *testing.T) {
	rule := NewLoggingNilCheckRule()

	assert.Equal(t, "logging-nil-check", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
