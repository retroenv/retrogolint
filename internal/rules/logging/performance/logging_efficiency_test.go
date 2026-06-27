package loggingperformance

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingEfficiencyRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "eager call in log field",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Process", log.String("result", expensiveCall()))
}`,
			wantViolations: 1,
		},
		{
			name: "lazy evaluation",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Process", log.StringFunc("result", func() string {
		return expensiveCall()
	}))
}`,
			wantViolations: 0,
		},
		{
			name: "no call expression",
			code: `package test
import "log"
func test(logger Logger, count int) {
	logger.Info("Process", log.Int("count", count))
}`,
			wantViolations: 0,
		},
		{
			name: "eager call in log wrapper field",
			code: `package test
import "log"
func test() {
	debugutil.LogOptimization(opts, Name,
		"Flattening struct",
		log.String("alloc", alloc.Name()),
		log.Int("field_count", len(fields)))
}`,
			wantViolations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingEfficiencyRule()
			violations := rule.Check(fset, file)

			if len(violations) != tt.wantViolations {
				for i, v := range violations {
					t.Logf("Violation %d: %s at %v", i+1, v.Message, v.Position)
				}
			}
			assert.Len(t, violations, tt.wantViolations)

			for _, v := range violations {
				assert.Equal(t, rule.Name(), v.Rule)
				assert.Equal(t, violation.SeverityInfo, v.Severity)
			}
		})
	}
}

func TestLoggingEfficiencyRule_Properties(t *testing.T) {
	rule := NewLoggingEfficiencyRule()

	assert.Equal(t, "logging-efficiency", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityInfo, rule.Severity())
}
