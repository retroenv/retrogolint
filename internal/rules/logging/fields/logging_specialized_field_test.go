package loggingfields

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingSpecializedFieldRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
		messageSubstr  string
	}{
		{
			name: "parameter typed uint16",
			code: `package test
import "log"
type Logger interface {
	Debug(msg string, fields ...any)
}
func test(logger Logger, addr uint16) {
	logger.Debug("address", log.Int("address", int(addr)))
}`,
			wantViolations: 1,
			messageSubstr:  "log.Uint16",
		},
		{
			name: "var declaration typed uint8",
			code: `package test
import "log"
type Logger interface {
	Info(msg string, fields ...any)
}
func test(logger Logger) {
	var offset uint8
	logger.Info("offset", log.Uint("offset", uint(offset)))
}`,
			wantViolations: 1,
			messageSubstr:  "log.Uint8",
		},
		{
			name: "int16 converted to int",
			code: `package test
import "log"
type Logger interface {
	Info(msg string, fields ...any)
}
func test(logger Logger, value int16) {
	logger.Info("value", log.Int("value", int(value)))
}`,
			wantViolations: 1,
			messageSubstr:  "log.Int16",
		},
		{
			name: "specialized field already used",
			code: `package test
import "log"
type Logger interface {
	Info(msg string, fields ...any)
}
func test(logger Logger, value int16) {
	logger.Info("value", log.Int16("value", value))
}`,
			wantViolations: 0,
		},
		{
			name: "short declaration with function call returning uint16",
			code: `package test
import "log"
type Logger interface {
	Debug(msg string, fields ...any)
}
func getSize() uint16 {
	return 42
}
func test(logger Logger) {
	size := getSize()
	logger.Debug("size", log.Int("size", int(size)))
}`,
			wantViolations: 1,
			messageSubstr:  "log.Uint16",
		},
		{
			name: "conversion without explicit type",
			code: `package test
import "log"
type Logger interface {
	Info(msg string, fields ...any)
}
func test(logger Logger) {
	x := uint16(0)
	logger.Info("value", log.Int("value", int(x)))
}`,
			wantViolations: 1,
			messageSubstr:  "log.Uint16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingSpecializedFieldRule()
			violations := rule.Check(fset, file)

			if len(violations) != tt.wantViolations {
				for i, v := range violations {
					t.Logf("Violation %d: %s", i+1, v.Message)
				}
			}
			assert.Len(t, violations, tt.wantViolations)

			if tt.messageSubstr != "" && len(violations) > 0 {
				assert.Contains(t, violations[0].Message, tt.messageSubstr)
			}
		})
	}
}

func TestLoggingSpecializedFieldRule_Properties(t *testing.T) {
	rule := NewLoggingSpecializedFieldRule()

	assert.Equal(t, "logging-specialized-field", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
