package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestInterfaceMultilineRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
		wantLine       int
	}{
		{
			name: "single-line interface with one method",
			code: `package test
type Archive struct{}
type Device interface{ TransferState(*Archive) error }
`,
			wantViolations: 1,
			wantLine:       3,
		},
		{
			name: "single-line interface with multiple methods",
			code: `package test
type Device interface{ Run(); Stop() }
`,
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "single-line embedded interface",
			code: `package test
type Reader interface{ io.Reader }
`,
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "single-line interface type alias",
			code: `package test
type Device = interface{ Run() }
`,
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "single-line named constraint",
			code: `package test
type Number interface{ ~int | ~float64 }
`,
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "single-line interface declared inside a function",
			code: `package test
func f() {
	type Device interface{ Run() }
	var _ Device
}
`,
			wantViolations: 1,
			wantLine:       3,
		},
		{
			name: "multiple single-line interfaces in one declaration",
			code: `package test
type (
	Reader interface{ Read([]byte) (int, error) }
	Writer interface{ Write([]byte) (int, error) }
)
`,
			wantViolations: 2,
		},
		{
			name: "multiline interface with one method",
			code: `package test
type Archive struct{}
type Device interface {
	TransferState(*Archive) error
}
`,
			wantViolations: 0,
		},
		{
			name: "multiline interface with multiple methods",
			code: `package test
type Device interface {
	Run()
	Stop()
}
`,
			wantViolations: 0,
		},
		{
			name: "multiline named constraint",
			code: `package test
type Number interface {
	~int | ~float64
}
`,
			wantViolations: 0,
		},
		{
			name: "empty interface declaration",
			code: `package test
type Values interface{}
`,
			wantViolations: 0,
		},
		{
			name: "single-line struct declaration",
			code: `package test
type Point struct{ X, Y int }
`,
			wantViolations: 0,
		},
		{
			name: "inline type parameter constraint",
			code: `package test
type Box[T interface{ Size() int }] struct {
	value T
}
`,
			wantViolations: 0,
		},
		{
			name: "anonymous interface in function parameter",
			code: `package test
func run(r interface{ Run() }) {
	r.Run()
}
`,
			wantViolations: 0,
		},
	}

	rule := NewInterfaceMultilineRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.wantViolations)
			if tt.wantLine > 0 && len(violations) > 0 {
				assert.Equal(t, tt.wantLine, violations[0].Position.Line)
			}
		})
	}
}

func TestInterfaceMultilineRule_Metadata(t *testing.T) {
	rule := NewInterfaceMultilineRule()

	assert.Equal(t, "codequality-interface-multiline", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
