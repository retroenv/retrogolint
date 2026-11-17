package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestFuncLiteralMultilineRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
		wantLine       int
	}{
		{
			name: "inline body as a call argument",
			code: `package test

func run(serve func() error) {
	wrap(func() error { return serve() })
}
`,
			wantViolations: 1,
			wantLine:       4,
		},
		{
			name: "two inline literal arguments in one call",
			code: `package test

func run(serve func() error) {
	wrap(func() error { return serve() }, func() error { return serve() })
}
`,
			wantViolations: 2,
			wantLine:       4,
		},
		{
			name: "inline body in an assignment",
			code: `package test

func run(serve func() error) {
	fn := func() error { return serve() }
	_ = fn
}
`,
			wantViolations: 1,
			wantLine:       4,
		},
		{
			name: "inline bodies in one assignment",
			code: `package test

func run() {
	serve, stop := func() error { return nil }, func() {}
	_, _ = serve, stop
}
`,
			wantViolations: 1,
			wantLine:       4,
		},
		{
			name: "inline body in a variable declaration",
			code: `package test

var serve = func() error { return nil }
`,
			wantViolations: 1,
			wantLine:       3,
		},
		{
			name: "call with inline argument assigned to a variable",
			code: `package test

func run(serve func() error) {
	fn := wrap(func() error { return serve() })
	_ = fn
}
`,
			wantViolations: 1,
			wantLine:       4,
		},
		{
			name: "multiline body argument",
			code: `package test

func run(serve func() error) {
	wrap(func() error {
		return serve()
	})
}
`,
			wantViolations: 0,
		},
		{
			name: "multiline body in an assignment",
			code: `package test

func run(serve func() error) error {
	fn := func() error {
		return serve()
	}
	return fn()
}
`,
			wantViolations: 0,
		},
		{
			name: "empty body argument",
			code: `package test

func run() {
	wrap(func() {})
}
`,
			wantViolations: 0,
		},
		{
			name: "function literal in a return",
			code: `package test

func run(serve func() error) func() error {
	return func() error { return serve() }
}
`,
			wantViolations: 0,
		},
		{
			name: "deferred function literal call",
			code: `package test

func run(serve func() error) {
	defer func() { _ = serve() }()
}
`,
			wantViolations: 0,
		},
		{
			name: "function literal in a struct field",
			code: `package test

type holder struct {
	serve func() error
}

func run(serve func() error) {
	_ = holder{serve: func() error { return serve() }}
}
`,
			wantViolations: 0,
		},
		{
			// Parentheses must not hide a function literal from the rule.
			name: "parenthesized inline body as a call argument",
			code: `package test

func run(serve func() error) {
	wrap(((func() error { return serve() })))
}
`,
			wantViolations: 1,
			wantLine:       4,
		},
		{
			name: "parenthesized inline body in an assignment",
			code: `package test

func run(serve func() error) {
	fn := ((func() error { return serve() }))
	_ = fn
}
`,
			wantViolations: 1,
			wantLine:       4,
		},
		{
			name: "parenthesized inline body in a variable declaration",
			code: `package test

var serve = ((func() error { return nil }))
`,
			wantViolations: 1,
			wantLine:       3,
		},
		{
			name: "parenthesized multiline bodies",
			code: `package test

var serve = (func() error {
	return nil
})

func run() {
	fn := (func() error {
		return serve()
	})
	wrap((func() error {
		return fn()
	}))
}
`,
			wantViolations: 0,
		},
		{
			name: "parenthesized empty bodies",
			code: `package test

var serve = (func() {})

func run() {
	fn := (func() {})
	wrap((func() {}))
	_ = fn
}
`,
			wantViolations: 0,
		},
		{
			name: "nested calls with inline literals",
			code: `package test

func run(serve func() error) {
	wrap(func() { inner(func() { serve() }) })
}
`,
			wantViolations: 2,
			wantLine:       4,
		},
	}

	rule := NewFuncLiteralMultilineRule()

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

func TestFuncLiteralMultilineRule_Metadata(t *testing.T) {
	rule := NewFuncLiteralMultilineRule()
	assert.Equal(t, "codequality-func-literal-multiline", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}
