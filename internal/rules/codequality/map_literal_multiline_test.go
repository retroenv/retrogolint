package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestMapLiteralMultilineRule(t *testing.T) {
	tests := []struct {
		name string
		code string
		want int
	}{
		{name: "explicit map", code: `var m = map[string]int{"a": 1, "b": 2}`, want: 1},
		{name: "named map", code: `type M map[string]int; const a, b = "a", "b"; var m = M{a: 1, b: 2}`, want: 1},
		{name: "alias", code: `type M map[string]int; type N = M; var m = N{"a": 1, "b": 2}`, want: 1},
		{name: "generic map", code: `type M[K comparable, V any] map[K]V; var m = M[string, int]{"a": 1, "b": 2}`, want: 1},
		{name: "first entry", code: "var m = map[string]int{\"a\": 1,\n\"b\": 2,\n}", want: 1},
		{name: "shared entry line", code: "var m = map[string]int{\n\"a\": 1, \"b\": 2,\n}", want: 1},
		{name: "multiline", code: "var m = map[string]int{\n\"a\": 1,\n\"b\": 2,\n}"},
		{name: "empty", code: `var m = map[string]int{}`},
		{name: "single", code: `var m = map[string]int{"a": 1}`},
		{name: "struct", code: `type S struct { A, B int }; var s = S{A: 1, B: 2}`, want: 1},
		{name: "array", code: `var a = [2]int{0: 1, 1: 2}`},
		{name: "named slice", code: `type S []int; var s = S{0: 1, 1: 2}`},
		{name: "cycle", code: `type A = B; type B = A; var a = A{X: 1, Y: 2}`, want: 1},
		{name: "nested explicit map", code: `var m = []map[string]int{map[string]int{"a": 1, "b": 2}}`, want: 1},
		{name: "map in slice", code: `var m = []map[string]int{{"a": 1, "b": 2}}`, want: 1},
		{name: "map in array", code: `var m = [1]map[string]int{{"a": 1, "b": 2}}`, want: 1},
		{name: "map in map", code: `var m = map[string]map[string]int{"outer": {"a": 1, "b": 2}}`, want: 1},
		{name: "named map in slice", code: `type M map[string]int; var m = []M{{"a": 1, "b": 2}}`, want: 1},
		{name: "map through nested slices", code: `var m = [][]map[string]int{{{"a": 1, "b": 2}}}`, want: 1},
		{name: "map through nested maps", code: `var m = map[int]map[int]map[int]int{0: {0: {1: 1, 2: 2}}}`, want: 1},
		{name: "pointer to map", code: `var m = []*map[string]int{{"a": 1, "b": 2}}`, want: 1},
		{name: "multiline nested map", code: "var m = []map[string]int{{\n\"a\": 1,\n\"b\": 2,\n}}"},
		{name: "nested slice exempt", code: `var m = [][]int{{1, 2}}`},
		{name: "nested struct exempt", code: `type S struct { A, B int }; var s = []S{{A: 1, B: 2}}`},
		{name: "map key struct exempt", code: `type S struct { A, B int }; var m = map[S]int{{A: 1, B: 2}: 1}`},
	}
	rule := NewStructLiteralMultilineRule()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", "package test\n"+test.code, 0)
			assert.NoError(t, err)
			assert.Len(t, rule.Check(fset, file), test.want)
		})
	}
	assert.Equal(t, "codequality-struct-literal-multiline", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
