package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestMultiAssignRule_Check(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		{
			name: "field assignments",
			code: `m.lastWriteCycle, m.lastWriteValid = cycle, ok`,
			want: 1,
		},
		{
			name: "short declaration",
			code: `cycle, ok := uint64(0), false`,
			want: 1,
		},
		{
			name: "three assignments",
			code: `a, b, c = 1, 2, 3`,
			want: 1,
		},
		{
			name: "swap",
			code: `a, b = b, a`,
		},
		{
			name: "field swap",
			code: `s.front, s.back = s.back, s.front`,
		},
		{
			name: "rotation",
			code: `a, b, c = b, c, a`,
		},
		{
			name: "not a permutation",
			code: `a, b = b, c`,
			want: 1,
		},
		{
			name: "swap with calls",
			code: `a[next()], b[next()] = b[next()], a[next()]`,
			want: 1,
		},
		{
			name: "swap with receive",
			code: `a, b = <-b, a`,
			want: 1,
		},
		{
			name: "short declaration with swapped values",
			code: `a, b := b, a`,
			want: 1,
		},
		{
			name: "multiple calls",
			code: `a, b := first(), second()`,
			want: 1,
		},
		{
			name: "multiline assignment",
			code: `a, b = first(),
	second()`,
			want: 1,
		},
		{
			name: "loop initialization and update",
			code: `for a, b := 0, 1; a < b; a, b = a+1, b+1 {}`,
			want: 2,
		},
		{
			name: "nested assignments",
			code: `a, b := func() int {
	x, y := 1, 2
	return x + y
}, 0`,
			want: 2,
		},
		{
			name: "separate assignments",
			code: `m.lastWriteCycle = cycle
m.lastWriteValid = ok`,
		},
		{
			name: "function results",
			code: `cycle, ok := m.CPUCycles()`,
		},
		{
			name: "assigned function results",
			code: `m.lastWriteCycle, m.lastWriteValid = m.CPUCycles()`,
		},
		{
			name: "map lookup",
			code: `value, ok := values[key]`,
		},
		{
			name: "type assertion",
			code: `value, ok := item.(int)`,
		},
		{
			name: "channel receive",
			code: `value, ok := <-values`,
		},
		{
			name: "range clause",
			code: `for key, value := range values { use(key, value) }`,
		},
		{
			name: "single value with commas",
			code: `values := []int{1, 2, 3}`,
		},
	}

	rule := NewMultiAssignRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			code := "package test\nfunc example() {\n" + tt.code + "\n}\n"
			file, err := parser.ParseFile(fset, "test.go", code, 0)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.want)
			for _, found := range violations {
				assert.Equal(t, rule.Name(), found.Rule)
				assert.Equal(t, rule.Severity(), found.Severity)
				assert.Equal(t, "avoid assigning multiple explicit values in one statement", found.Message)
			}
		})
	}
}

func TestMultiAssignRule_Metadata(t *testing.T) {
	rule := NewMultiAssignRule()

	assert.Equal(t, "codequality-multi-assign", rule.Name())
	assert.Equal(t, "Avoid assignments with multiple explicit values", rule.Description())
	assert.Equal(t, "codequality", rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
