package rules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/rules/api"
)

func TestIsSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty string", "", true},
		{"lowercase", "test", true},
		{"lowercase with underscore", "test_name", true},
		{"lowercase with number", "test123", true},
		{"lowercase with underscore and number", "test_123", true},
		{"starts with underscore", "_test", true},
		{"multiple underscores", "test__name", true},
		{"camelCase", "testName", false},
		{"PascalCase", "TestName", false},
		{"starts with uppercase", "Test", false},
		{"contains uppercase", "test_Name", false},
		{"contains space", "test name", false},
		{"contains hyphen", "test-name", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := api.IsSnakeCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"already snake_case", "test_name", "test_name"},
		{"camelCase", "testName", "test_name"},
		{"PascalCase", "TestName", "test_name"},
		{"multiple capitals", "testHTTPServer", "test_h_t_t_p_server"},
		{"consecutive capitals", "HTTPServer", "h_t_t_p_server"},
		{"single letter", "a", "a"},
		{"uppercase letter", "A", "a"},
		{"with numbers", "test123Name", "test123_name"},
		{"starts with number", "5test", "5test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := api.ToSnakeCase(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractStringLiteral(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantStr string
		wantOk  bool
	}{
		{
			name:    "double quoted string",
			code:    `package test; var s = "hello"`,
			wantStr: "hello",
			wantOk:  true,
		},
		{
			name:    "backtick string",
			code:    "package test; var s = `hello`",
			wantStr: "hello",
			wantOk:  true,
		},
		{
			name:    "empty string",
			code:    `package test; var s = ""`,
			wantStr: "",
			wantOk:  true,
		},
		{
			name:    "string with spaces",
			code:    `package test; var s = "hello world"`,
			wantStr: "hello world",
			wantOk:  true,
		},
		{
			name:    "escaped string",
			code:    `package test; var s = "hello\nworld"`,
			wantStr: "hello\nworld",
			wantOk:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			// Find the string literal in the AST
			var gotStr string
			var gotOk bool
			for _, decl := range file.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					for _, spec := range genDecl.Specs {
						if valueSpec, ok := spec.(*ast.ValueSpec); ok && len(valueSpec.Values) > 0 {
							gotStr, gotOk = api.ExtractStringLiteral(valueSpec.Values[0])
						}
					}
				}
			}

			assert.Equal(t, tt.wantOk, gotOk)
			if gotOk {
				assert.Equal(t, tt.wantStr, gotStr)
			}
		})
	}
}

func TestIsLoggerMethod(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "Info method",
			code: `package test; func test(l Logger) { l.Info("test") }`,
			want: true,
		},
		{
			name: "Debug method",
			code: `package test; func test(l Logger) { l.Debug("test") }`,
			want: true,
		},
		{
			name: "Error method",
			code: `package test; func test(l Logger) { l.Error("test") }`,
			want: true,
		},
		{
			name: "Infof method",
			code: `package test; func test(l Logger) { l.Infof("test %s", "msg") }`,
			want: true,
		},
		{
			name: "non-logger method",
			code: `package test; func test(l Logger) { l.Other("test") }`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			// Find the call expression
			var got bool
			inspectCalls(file, func(call *ast.CallExpr) {
				if api.IsLoggerMethod(call) {
					got = true
				}
			})

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsLogFieldCall(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "String field",
			code: `package test; import "log"; func test() { log.String("key", "val") }`,
			want: true,
		},
		{
			name: "Int field",
			code: `package test; import "log"; func test() { log.Int("key", 5) }`,
			want: true,
		},
		{
			name: "non-field method",
			code: `package test; import "log"; func test() { log.Other("key", "val") }`,
			want: false,
		},
		{
			name: "selector on non-log package",
			code: `package test; func test(instr Instr) { instr.String() }`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			// Find the call expression
			var got bool
			inspectCalls(file, func(call *ast.CallExpr) {
				if api.IsLogFieldCall(call) {
					got = true
				}
			})

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistryValidateFilters(t *testing.T) {
	registry := NewRegistry()

	assert.NoError(t, registry.ValidateFilters([]string{
		"logging",
		"logging-capitalization",
		"codequality",
	}))

	err := registry.ValidateFilters([]string{"logging-capitalisation"})
	assert.Error(t, err)
}

// Helper function to inspect calls in an AST.
func inspectCalls(file *ast.File, fn func(*ast.CallExpr)) {
	ast.Inspect(file, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			fn(call)
		}
		return true
	})
}
