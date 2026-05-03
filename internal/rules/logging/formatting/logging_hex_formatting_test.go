package loggingformatting

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingHexFormattingRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "fmt.Sprintf hex formatting",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, address uint16) {
	logger.Debug("Handler", log.String("addr", fmt.Sprintf("0x%04X", address)))
}`,
			wantViolations: 1,
		},
		{
			name: "fmt.Sprintf non-hex formatting",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, address uint16) {
	logger.Debug("Handler", log.String("addr", fmt.Sprintf("%d", address)))
}`,
			wantViolations: 0,
		},
		{
			name: "log.Hex field",
			code: `package test
import "log"
func test(logger Logger, address uint16) {
	logger.Debug("Handler", log.Hex("addr", address))
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingHexFormattingRule()
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

func TestLoggingHexFormattingRule_Properties(t *testing.T) {
	rule := NewLoggingHexFormattingRule()

	assert.Equal(t, "logging-hex-formatting", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
