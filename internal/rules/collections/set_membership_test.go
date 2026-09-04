package collections

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestSetMembershipRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "set scan returns on match",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func contains(items set.Set[string], target string) bool {
	for item := range items {
		if item == target {
			return true
		}
	}
	return false
}`,
			wantViolations: 1,
		},
		{
			name: "set scan returns error on match",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func validate(items set.Set[string], target string) error {
	for item := range items {
		if target == item {
			return errDuplicate
		}
	}
	return nil
}`,
			wantViolations: 1,
		},
		{
			name: "set iteration performs other work",
			code: `package test
import "github.com/retroenv/retrogolib/set"
func write(items set.Set[string]) {
	for item := range items {
		println(item)
	}
}`,
			wantViolations: 0,
		},
		{
			name: "map scan without set import",
			code: `package test
func contains(items map[string]struct{}, target string) bool {
	for item := range items {
		if item == target {
			return true
		}
	}
	return false
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewSetMembershipRule()
			violations := rule.Check(fset, file)

			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestSetMembershipRule_Properties(t *testing.T) {
	rule := NewSetMembershipRule()

	assert.Equal(t, "collections-set-membership", rule.Name())
	assert.Equal(t, api.CategoryCollections, rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}
