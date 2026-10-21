package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestInlineInterfaceRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "inline interface in type assertion",
			code: `package arch
func f(v any) {
	resolver, ok := v.(interface{ Resolve(uint16) (int, bool) })
	_, _ = resolver, ok
}
`,
			wantViolations: 1,
		},
		{
			name: "named interface in type assertion",
			code: `package arch
type resolver interface{ Resolve(uint16) (int, bool) }
func f(v any) {
	r, ok := v.(resolver)
	_, _ = r, ok
}
`,
			wantViolations: 0,
		},
		{
			name: "named interface declaration",
			code: `package arch
type Resolver interface{ Resolve(uint16) (int, bool) }
`,
			wantViolations: 0,
		},
		{
			name: "interface type alias declaration",
			code: `package arch
type Resolver = interface{ Resolve(uint16) (int, bool) }
`,
			wantViolations: 0,
		},
		{
			name: "empty interface",
			code: `package arch
func f(v interface{}) {
	_ = v
}
`,
			wantViolations: 0,
		},
		{
			name: "empty interface in map value",
			code: `package arch
var values map[string]interface{}
`,
			wantViolations: 0,
		},
		{
			name: "inline interface in function parameter",
			code: `package arch
func f(v interface{ Run() }) {
	v.Run()
}
`,
			wantViolations: 1,
		},
		{
			name: "inline interface in function result",
			code: `package arch
func f() interface{ Run() } {
	return nil
}
`,
			wantViolations: 1,
		},
		{
			name: "inline interface in struct field",
			code: `package arch
type Service struct {
	runner interface{ Run() }
}
`,
			wantViolations: 1,
		},
		{
			name: "inline interface in variable declaration",
			code: `package arch
var runner interface{ Run() }
`,
			wantViolations: 1,
		},
		{
			name: "inline embedded interface",
			code: `package arch
var reader interface{ io.Reader }
`,
			wantViolations: 1,
		},
		{
			name: "inline interface inside named interface method",
			code: `package arch
type Visitor interface{ Visit(interface{ Accept() }) }
`,
			wantViolations: 1,
		},
		{
			name: "type parameter constraint on function",
			code: `package arch
func Max[T interface{ ~int | ~uint16 }](a, b T) T {
	if a > b {
		return a
	}
	return b
}
`,
			wantViolations: 0,
		},
		{
			name: "type parameter constraint on type",
			code: `package arch
type Box[T interface{ Size() int }] struct {
	value T
}
`,
			wantViolations: 0,
		},
		{
			name: "inline interface in function parameter of generic function",
			code: `package arch
func Run[T any](v T, runner interface{ Run() }) {
	runner.Run()
}
`,
			wantViolations: 1,
		},
		{
			name: "multiple inline interfaces",
			code: `package arch
func f(a interface{ Run() }, b interface{ Stop() }) {
	_, _ = a, b
}
`,
			wantViolations: 2,
		},
	}

	rule := NewInlineInterfaceRule()

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

func TestInlineInterfaceRule_Metadata(t *testing.T) {
	rule := NewInlineInterfaceRule()

	assert.Equal(t, "codequality-inline-interface", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
