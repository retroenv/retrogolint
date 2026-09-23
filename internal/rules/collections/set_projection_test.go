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
			name: "unrelated method with set import",
			code: `package test
import "github.com/retroenv/retrogolib/set"
type values struct{}
func (values) ToSlice() []string { return nil }
func collect(items values, members set.Set[string]) []string {
	_ = members
	return items.ToSlice()
}`,
			wantViolations: 0,
		},
		{
			name: "constructed set",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func values() []string {
	items := set.New[string]()
	return items.ToSlice()
}`,
			wantViolations: 1,
		},
		{
			name: "set fields and function results",
			code: `package test
import "github.com/retroenv/retrogolib/set"
type holder struct { members set.Set[string] }
func makeMembers() set.Set[string] { return set.New[string]() }
func (h holder) Members() set.Set[string] { return h.members }
func project(h holder) {
	_ = h.members.ToSlice()
	_ = makeMembers().ToSlice()
	_ = h.Members().ToSlice()
}`,
			wantViolations: 3,
		},
		{
			name: "unrelated fields and function results",
			code: `package test
import "github.com/retroenv/retrogolib/set"
type values struct{}
func (values) ToSlice() []string { return nil }
type holder struct { members values }
func makeValues() values { return values{} }
func (h holder) Values() values { return h.members }
func project(h holder, members set.Set[string]) {
	_ = members
	_ = h.members.ToSlice()
	_ = makeValues().ToSlice()
	_ = h.Values().ToSlice()
}`,
			wantViolations: 0,
		},
		{
			name: "shadowed set parameter",
			code: `package test
import "github.com/retroenv/retrogolib/set"
type values struct{}
func (values) ToSlice() []string { return nil }
func collect(items set.Set[string]) []string {
	if true {
		items := values{}
		return items.ToSlice()
	}
	return items.ToSlice()
}`,
			wantViolations: 1,
		},
		{
			name: "set declarations and copies",
			code: `package test
import collection "github.com/retroenv/retrogolib/set"
type members = collection.Set[string]
func collect(input *members) {
	var declared collection.Set[string]
	var initialized = collection.NewFromSlice([]string{"a"})
	created := make(collection.Set[string])
	alias := created
	_ = declared.ToSlice()
	_ = initialized.ToSlice()
	_ = alias.Copy().ToSlice()
	_ = (*input).ToSlice()
	_ = (collection.Set[string]{}).ToSlice()
	_ = collection.Set[string](nil).ToSlice()
}`,
			wantViolations: 6,
		},
		{
			name: "unknown and non-set constructors",
			code: `package test
import "github.com/retroenv/retrogolib/set"
type values struct{}
func (values) ToSlice() []string { return nil }
func collect(items set.Set[string]) {
	var unrelated values
	local := values{}
	_ = unrelated.ToSlice()
	_ = local.ToSlice()
	_ = unknown.ToSlice()
	_ = factory().ToSlice()
	_ = items.Unrelated().ToSlice()
}`,
			wantViolations: 0,
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
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments|parser.SkipObjectResolution)
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
