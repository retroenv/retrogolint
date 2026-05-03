package loggingformatting

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestLoggingFmtInFieldRule(t *testing.T) {
	tests := []struct {
		name            string
		code            string
		wantViolations  int
		wantMsgContains []string
	}{
		{
			name: "fmt.Sprintf with %v in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, val any) {
	logger.Info("Test", log.String("value", fmt.Sprintf("%v", val)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Any()"},
		},
		{
			name: "fmt.Sprintf with %s in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, s string) {
	logger.Info("Test", log.String("value", fmt.Sprintf("%s", s)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "string value directly"},
		},
		{
			name: "fmt.Sprintf with %d in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, n int) {
	logger.Info("Test", log.String("value", fmt.Sprintf("%d", n)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Int()"},
		},
		{
			name: "fmt.Sprintf with %x in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, addr uint16) {
	logger.Info("Test", log.String("addr", fmt.Sprintf("0x%04X", addr)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Hex()"},
		},
		{
			name: "fmt.Sprintf with %T in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, val any) {
	logger.Info("Test", log.String("type", fmt.Sprintf("%T", val)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Type()"},
		},
		{
			name: "fmt.Sprintf with %p in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, ptr *int) {
	logger.Info("Test", log.String("ptr", fmt.Sprintf("%p", ptr)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Uintptr()"},
		},
		{
			name: "fmt.Sprintf with %f in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, val float64) {
	logger.Info("Test", log.String("metric", fmt.Sprintf("%.2f", val)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Float64()"},
		},
		{
			name: "fmt.Sprintf with %t in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, b bool) {
	logger.Info("Test", log.String("flag", fmt.Sprintf("%t", b)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "log.Bool()"},
		},
		{
			name: "fmt.Sprintf with complex format string",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, a int, b string) {
	logger.Info("Test", log.String("pattern", fmt.Sprintf("%d-bit %s rotation", a, b)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf", "separate structured log fields"},
		},
		{
			name: "fmt.Sprint in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, val any) {
	logger.Info("Test", log.String("value", fmt.Sprint(val)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprint", "log.Any()"},
		},
		{
			name: "fmt.Sprintln in log.String",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, val any) {
	logger.Info("Test", log.String("value", fmt.Sprintln(val)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintln", "log.Any()"},
		},
		{
			name: "fmt.Sprintf in log.Int",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, n int) {
	logger.Info("Test", log.Int("count", fmt.Sprintf("%d", n)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf"},
		},
		{
			name: "fmt.Sprintf in log.Uint64",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, val uint64) {
	logger.Info("Test", log.Uint64("value", fmt.Sprintf("%d", val)))
}`,
			wantViolations:  1,
			wantMsgContains: []string{"Sprintf"},
		},
		{
			name: "multiple violations in same log call",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger, a int, b string) {
	logger.Info("Test",
		log.String("val1", fmt.Sprintf("%v", a)),
		log.String("val2", fmt.Sprintf("%v", b)))
}`,
			wantViolations: 2,
		},
		{
			name: "valid log.String without fmt",
			code: `package test
import "log"
func test(logger Logger, s string) {
	logger.Info("Test", log.String("value", s))
}`,
			wantViolations: 0,
		},
		{
			name: "valid log.Int without fmt",
			code: `package test
import "log"
func test(logger Logger, n int) {
	logger.Info("Test", log.Int("count", n))
}`,
			wantViolations: 0,
		},
		{
			name: "valid log.Float64 without fmt",
			code: `package test
import "log"
func test(logger Logger, val float64) {
	logger.Info("Test", log.Float64("value", val))
}`,
			wantViolations: 0,
		},
		{
			name: "valid log.Hex",
			code: `package test
import "log"
func test(logger Logger, addr uint16) {
	logger.Info("Test", log.Hex("addr", addr))
}`,
			wantViolations: 0,
		},
		{
			name: "fmt.Sprintf in non-log context",
			code: `package test
import "fmt"
func test() {
	s := fmt.Sprintf("%v", 42)
	_ = s
}`,
			wantViolations: 0,
		},
		{
			name: "fmt package called but not in log field",
			code: `package test
import (
	"fmt"
	"log"
)
func test(logger Logger) {
	msg := fmt.Sprintf("test")
	logger.Info(msg)
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingFmtInFieldRule()
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

				// Check that the message contains expected substrings
				for _, substr := range tt.wantMsgContains {
					assert.True(t, strings.Contains(v.Message, substr),
						"Expected message to contain %q, got: %s", substr, v.Message)
				}
			}
		})
	}
}

func TestLoggingFmtInFieldRule_Properties(t *testing.T) {
	rule := NewLoggingFmtInFieldRule()

	assert.Equal(t, "logging-fmt-in-field", rule.Name())
	assert.Equal(t, api.CategoryLogging, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
