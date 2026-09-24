package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestStructLiteralMultilineRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
		wantLine       int
	}{
		{
			name: "single-line struct literal with multiple fields",
			code: `package test
type CPU struct {
	Core any
	memory any
	fetch any
}
func New(memory, fetch any) *CPU {
	return &CPU{Core: new(any), memory: memory, fetch: fetch}
}
`,
			wantViolations: 1,
			wantLine:       8,
		},
		{
			name: "multiline struct literal with multiple fields",
			code: `package test
type CPU struct {
	Core any
	memory any
	fetch any
}
func New(memory, fetch any) *CPU {
	return &CPU{
		Core: new(any),
		memory: memory,
		fetch: fetch,
	}
}
`,
			wantViolations: 0,
		},
		{
			name: "multiline struct literal with fields on shared lines",
			code: `package test
type CPU struct {
	Core any
	memory any
	fetch any
}
func New(memory, fetch any) *CPU {
	return &CPU{
		Core: new(any), memory: memory,
		fetch: fetch,
	}
}
`,
			wantViolations: 1,
			wantLine:       9,
		},
		{
			name: "nested multiline struct literal with fields on shared lines",
			code: `package test
type Memory struct {
	BasicMemory any
	Size any
}
type CPU struct {
	Core any
	memory any
}
func New(memory any) *CPU {
	return &CPU{
		Core: &Memory{
			BasicMemory: memory, Size: memory,
		},
		memory: memory,
	}
}
`,
			wantViolations: 1,
		},
		{
			name: "single-field struct literal",
			code: `package test
type Memory struct { BasicMemory any }
func New(memory any) *Memory {
	return &Memory{BasicMemory: memory}
}
`,
			wantViolations: 0,
		},
		{
			name: "nested single-field struct literal",
			code: `package test
type Memory struct { BasicMemory any }
type CPU struct {
	Core any
	memory any
}
func New(memory any) *CPU {
	return &CPU{
		Core: &Memory{BasicMemory: memory},
		memory: memory,
	}
}
`,
			wantViolations: 0,
		},
		{
			name: "explicit map literal",
			code: `package test
var values = map[string]int{"one": 1, "two": 2}
`,
			wantViolations: 0,
		},
		{
			name: "slice literal",
			code: `package test
var values = []int{1, 2}
`,
			wantViolations: 0,
		},
		{
			name: "positional struct literal",
			code: `package test
type Point struct { X, Y int }
var point = Point{1, 2}
`,
			wantViolations: 0,
		},
		{
			name: "generic struct literal",
			code: `package test
type Pair[T any] struct { Left, Right T }
var pair = Pair[int]{Left: 1, Right: 2}
`,
			wantViolations: 1,
		},
	}

	rule := NewStructLiteralMultilineRule()

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

func TestStructLiteralMultilineRule_Metadata(t *testing.T) {
	rule := NewStructLiteralMultilineRule()

	assert.Equal(t, "codequality-struct-literal-multiline", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
