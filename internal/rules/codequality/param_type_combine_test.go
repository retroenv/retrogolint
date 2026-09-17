package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestParamTypeCombineRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "multiline consecutive parameters",
			code: `package test
func rewrite(nodes []Node, start int, finalLoad int, finalStore int, addr Node,
	finalStoreInstr Instruction) {}
`,
			wantViolations: 1,
		},
		{
			name: "single-line consecutive parameters",
			code: `package test
func rewrite(start int, finalLoad int, finalStore int) {}
`,
			wantViolations: 1,
		},
		{
			name: "multiple runs",
			code: `package test
func rewrite(start int, finalLoad int, source Node, destination Node) {}
`,
			wantViolations: 2,
		},
		{
			name: "already combined",
			code: `package test
func rewrite(start, finalLoad, finalStore int) {}
`,
			wantViolations: 0,
		},
		{
			name: "same type is not consecutive",
			code: `package test
func rewrite(start int, node Node, finalLoad int) {}
`,
			wantViolations: 0,
		},
		{
			name: "different complex types",
			code: `package test
func rewrite(left []Node, right []*Node) {}
`,
			wantViolations: 0,
		},
		{
			name: "same complex type",
			code: `package test
func rewrite(left map[string][]Node, right map[string][]Node) {}
`,
			wantViolations: 1,
		},
		{
			name: "method parameters",
			code: `package test
type optimizer struct{}
func (optimizer) rewrite(start int, finalLoad int) {}
`,
			wantViolations: 1,
		},
		{
			name: "signature comment",
			code: `package test
func rewrite(start int,
	// Keep the load separate for documentation.
	finalLoad int) {}
`,
			wantViolations: 0,
		},
	}

	rule := NewParamTypeCombineRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestParamTypeCombineRule_Metadata(t *testing.T) {
	rule := NewParamTypeCombineRule()

	assert.Equal(t, "codequality-param-type-combine", rule.Name())
	assert.Equal(t, "Combine consecutive function parameters that have the same type", rule.Description())
	assert.Equal(t, "codequality", rule.Category())
}
