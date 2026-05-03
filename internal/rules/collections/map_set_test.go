package collections

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestMapSetRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "map int bool",
			code: `package test
func test() {
	var ints map[int]bool
	_ = ints
}`,
			wantViolations: 1,
		},
		{
			name: "map inline struct bool",
			code: `package test
func test() {
	var value struct{ id int }
	var set map[struct{ id int }]bool
	_ = value
	_ = set
}`,
			wantViolations: 1,
		},
		{
			name: "map struct type bool",
			code: `package test
type key struct{ id int }

func test() {
	var set map[key]bool
	_ = set
}`,
			wantViolations: 1,
		},
		{
			name: "map string bool",
			code: `package test
func test() {
	var values map[string]bool
	_ = values
}`,
			wantViolations: 1,
		},
		{
			name: "map int empty struct",
			code: `package test
func test() {
	var visited map[int]struct{}
	_ = visited
}`,
			wantViolations: 1,
		},
		{
			name: "map string empty struct",
			code: `package test
func test() {
	var set map[string]struct{}
	_ = set
}`,
			wantViolations: 1,
		},
		{
			name: "map custom type empty struct",
			code: `package test
type ID string

func test() {
	var ids map[ID]struct{}
	_ = ids
}`,
			wantViolations: 1,
		},
		{
			name: "map interface empty struct",
			code: `package test
func test() {
	var values map[any]struct{}
	_ = values
}`,
			wantViolations: 1,
		},
		{
			name: "map non-empty struct - no violation",
			code: `package test
func test() {
	var values map[string]struct{ count int }
	_ = values
}`,
			wantViolations: 0,
		},
		{
			name: "map string int - no violation",
			code: `package test
func test() {
	var counts map[string]int
	_ = counts
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewMapSetRule()
			violations := rule.Check(fset, file)

			if len(violations) != tt.wantViolations {
				for i, v := range violations {
					t.Logf("Violation %d: %s", i+1, v.Message)
				}
			}
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestMapSetRule_Properties(t *testing.T) {
	rule := NewMapSetRule()

	assert.Equal(t, "collections-map-set", rule.Name())
	assert.Equal(t, api.CategoryCollections, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
