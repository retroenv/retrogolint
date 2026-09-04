package collections

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestSetProjectionRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "set to slice",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func values(items set.Set[string]) []string {
	return items.ToSlice()
}`,
			wantViolations: 1,
		},
		{
			name: "set alias to slice",
			code: `package test
import collection "github.com/retroenv/retrogolib/set"
func values(items collection.Set[string]) []string {
	return items.ToSlice()
}`,
			wantViolations: 1,
		},
		{
			name: "sorted projection",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func values(items set.Set[string]) []string {
	return set.Sorted(items)
}`,
			wantViolations: 0,
		},
		{
			name: "unrelated to slice method",
			code: `package test
type values struct{}
func (values) ToSlice() []string { return nil }
func collect(items values) []string {
	return items.ToSlice()
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewSetProjectionRule()
			violations := rule.Check(fset, file)

			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestSetProjectionRule_Properties(t *testing.T) {
	rule := NewSetProjectionRule()

	assert.Equal(t, "collections-set-projection", rule.Name())
	assert.Equal(t, api.CategoryCollections, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
