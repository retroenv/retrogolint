package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestParamPriorityRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "correct order context then logger then other",
			code: `package test
import (
	"context"
	"log/slog"
)
func Foo(ctx context.Context, logger *slog.Logger, name string) {}
`,
			wantViolations: 0,
		},
		{
			name: "correct order context then other",
			code: `package test
import "context"
func Foo(ctx context.Context, name string) {}
`,
			wantViolations: 0,
		},
		{
			name: "correct order logger then other",
			code: `package test
import "log/slog"
func Foo(logger *slog.Logger, name string) {}
`,
			wantViolations: 0,
		},
		{
			name: "logger before context",
			code: `package test
import (
	"context"
	"log/slog"
)
func Foo(logger *slog.Logger, ctx context.Context, name string) {}
`,
			wantViolations: 1,
		},
		{
			name: "other before context",
			code: `package test
import "context"
func Foo(name string, ctx context.Context) {}
`,
			wantViolations: 1,
		},
		{
			name: "other before logger",
			code: `package test
import "log/slog"
func Foo(name string, logger *slog.Logger) {}
`,
			wantViolations: 1,
		},
		{
			name: "single param no violation",
			code: `package test
func Foo(name string) {}
`,
			wantViolations: 0,
		},
		{
			name: "no params no violation",
			code: `package test
func Foo() {}
`,
			wantViolations: 0,
		},
		{
			name: "method with correct order",
			code: `package test
import (
	"context"
	"log/slog"
)
type S struct{}
func (s *S) Foo(ctx context.Context, logger *slog.Logger, name string) {}
`,
			wantViolations: 0,
		},
		{
			name: "method with wrong order",
			code: `package test
import (
	"context"
	"log/slog"
)
type S struct{}
func (s *S) Foo(logger *slog.Logger, ctx context.Context) {}
`,
			wantViolations: 1,
		},
		{
			name: "multiple functions one wrong",
			code: `package test
import (
	"context"
	"log/slog"
)
func Good(ctx context.Context, logger *slog.Logger) {}
func Bad(logger *slog.Logger, ctx context.Context) {}
`,
			wantViolations: 1,
		},
		{
			name: "context after other params",
			code: `package test
import (
	"context"
	"log/slog"
)
func Bad(name string, count int, ctx context.Context, logger *slog.Logger) {}
`,
			wantViolations: 1,
		},
	}

	rule := NewParamPriorityRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestParamPriorityRule_Metadata(t *testing.T) {
	rule := NewParamPriorityRule()

	assert.Equal(t, "codequality-param-priority", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
