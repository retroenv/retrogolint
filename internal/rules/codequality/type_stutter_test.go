package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestTypeStutterRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "stutter with suffix",
			code: `package register
type RegisterMeta struct{}
`,
			wantViolations: 1,
		},
		{
			name: "exact match capitalized is not stutter",
			code: `package register
type Register struct{}
`,
			wantViolations: 0,
		},
		{
			name: "no stutter different prefix",
			code: `package register
type Meta struct{}
`,
			wantViolations: 0,
		},
		{
			name: "coincidental prefix lowercase boundary",
			code: `package register
type Registered struct{}
`,
			wantViolations: 0,
		},
		{
			name: "unexported type not flagged",
			code: `package register
type registerMeta struct{}
`,
			wantViolations: 0,
		},
		{
			name: "external test package stripped",
			code: `package register_test
type RegisterMeta struct{}
`,
			wantViolations: 1,
		},
		{
			name: "short package name skipped",
			code: `package x
type XFoo struct{}
`,
			wantViolations: 0,
		},
		{
			name: "multiple types one stutters",
			code: `package store
type StoreItem struct{}
type Item struct{}
`,
			wantViolations: 1,
		},
		{
			name: "interface stutter",
			code: `package handler
type HandlerFunc interface{}
`,
			wantViolations: 1,
		},
		{
			name: "type alias stutter",
			code: `package config
type ConfigMap = map[string]string
`,
			wantViolations: 1,
		},
	}

	rule := NewTypeStutterRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestTypeStutterRule_Metadata(t *testing.T) {
	rule := NewTypeStutterRule()

	assert.Equal(t, "codequality-type-stutter", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
