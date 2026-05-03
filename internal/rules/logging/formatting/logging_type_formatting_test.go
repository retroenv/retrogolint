package loggingformatting

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingTypeFormattingRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "fmt.Sprintf type formatting",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, addr any) {
	logger.Debug("Handler", log.String("addr_type", fmt.Sprintf("%T", addr)))
}`,
			wantViolations: 1,
		},
		{
			name: "fmt.Sprintf non-type formatting",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, value int) {
	logger.Debug("Handler", log.String("value", fmt.Sprintf("%d", value)))
}`,
			wantViolations: 0,
		},
		{
			name: "log.Type field",
			code: `package test
import "log"
func test(logger Logger, addr any) {
	logger.Debug("Handler", log.Type("addr_type", addr))
}`,
			wantViolations: 0,
		},
		{
			name: "multiple %T in format string",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, a, b any) {
	logger.Debug("Types", log.String("types", fmt.Sprintf("%T and %T", a, b)))
}`,
			wantViolations: 1,
		},
		{
			name: "regular log.String usage",
			code: `package test
import "log"
func test(logger Logger) {
	logger.Info("Test", log.String("key", "value"))
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingTypeFormattingRule()
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

func TestLoggingTypeFormattingRule_Properties(t *testing.T) {
	rule := NewLoggingTypeFormattingRule()

	assert.Equal(t, "logging-type-formatting", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
