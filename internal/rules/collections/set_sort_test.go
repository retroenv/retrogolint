package collections

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestSetSortRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "manual set collection and sort",
			code: `package test
import (
	"slices"
	"github.com/retroenv/retrogolib/set"
)
func values(items set.Set[string]) []string {
	result := make([]string, 0, len(items))
	for item := range items {
		result = append(result, item)
	}
	slices.Sort(result)
	return result
}`,
			wantViolations: 1,
		},
		{
			name: "unsorted collection",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func values(items set.Set[string]) []string {
	var result []string
	for item := range items {
		result = append(result, item)
	}
	return result
}`,
			wantViolations: 0,
		},
		{
			name: "sort without set import",
			code: `package test
import "slices"
func values(items map[string]struct{}) []string {
	var result []string
	for item := range items {
		result = append(result, item)
	}
	slices.Sort(result)
	return result
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewSetSortRule()
			violations := rule.Check(fset, file)

			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestSetSortRule_Properties(t *testing.T) {
	rule := NewSetSortRule()

	assert.Equal(t, "collections-set-sort", rule.Name())
	assert.Equal(t, api.CategoryCollections, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
