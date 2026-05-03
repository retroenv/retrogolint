package api

import (
	"go/ast"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func loggerCall(receiver, method string) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: receiver},
			Sel: &ast.Ident{Name: method},
		},
	}
}

func TestIsLoggerMethod(t *testing.T) {
	for _, method := range []string{"Debug", "Info", "Warn", "Error", "Fatal", "Panic",
		"Debugf", "Infof", "Warnf", "Errorf", "Fatalf", "Panicf"} {
		assert.True(t, IsLoggerMethod(loggerCall("logger", method)), method)
	}
	for _, excluded := range []string{"fmt", "errors", "log", "t", "b"} {
		assert.False(t, IsLoggerMethod(loggerCall(excluded, "Info")), excluded)
	}
	assert.False(t, IsLoggerMethod(loggerCall("logger", "Println")))
	assert.False(t, IsLoggerMethod(&ast.CallExpr{Fun: &ast.Ident{Name: "Info"}}))
}

func TestIsLogFieldCall(t *testing.T) {
	for _, method := range []string{"String", "Int", "Int64", "Uint", "Uint64", "Float64",
		"Bool", "Duration", "Time", "Error", "Stringer", "Hex", "Type", "StringFunc", "IntFunc"} {
		call := &ast.CallExpr{
			Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "log"}, Sel: &ast.Ident{Name: method}},
		}
		assert.True(t, IsLogFieldCall(call), method)
	}
	call := &ast.CallExpr{
		Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "log"}, Sel: &ast.Ident{Name: "Unknown"}},
	}
	assert.False(t, IsLogFieldCall(call))
	assert.False(t, IsLogFieldCall(&ast.CallExpr{Fun: &ast.Ident{Name: "String"}}))
}

func TestIsSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"", true},
		{"foo", true},
		{"foo_bar", true},
		{"foo_bar_123", true},
		{"fooBar", false},
		{"FooBar", false},
		{"_foo", true},
		{"FOO", false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, IsSnakeCase(tt.input), tt.input)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"foo", "foo"},
		{"fooBar", "foo_bar"},
		{"FooBar", "foo_bar"},
		{"HTTPServer", "h_t_t_p_server"},
		{"myFieldName", "my_field_name"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, ToSnakeCase(tt.input), tt.input)
	}
}

func TestContainsFormatVerb(t *testing.T) {
	assert.True(t, ContainsFormatVerb("%s", 's'))
	assert.True(t, ContainsFormatVerb("value: %d", 'd'))
	assert.False(t, ContainsFormatVerb("%s", 'd'))
	assert.False(t, ContainsFormatVerb("%%s", 's'))
	assert.False(t, ContainsFormatVerb("no verbs", 's'))
	assert.True(t, ContainsFormatVerb("%v %s", 's'))
}

func TestExtractStringLiteral(t *testing.T) {
	lit := &ast.BasicLit{Kind: token.STRING, Value: `"hello"`}
	val, ok := ExtractStringLiteral(lit)
	assert.True(t, ok)
	assert.Equal(t, "hello", val)

	_, ok = ExtractStringLiteral(&ast.BasicLit{Kind: token.INT, Value: "42"})
	assert.False(t, ok)

	_, ok = ExtractStringLiteral(&ast.BasicLit{Kind: token.STRING, Value: `"unterminated`})
	assert.False(t, ok)

	_, ok = ExtractStringLiteral(&ast.Ident{Name: "x"})
	assert.False(t, ok)
}

func TestExprToString(t *testing.T) {
	assert.Equal(t, "foo", ExprToString(&ast.Ident{Name: "foo"}))
	assert.Equal(t, "pkg.Field", ExprToString(&ast.SelectorExpr{
		X:   &ast.Ident{Name: "pkg"},
		Sel: &ast.Ident{Name: "Field"},
	}))
	assert.Equal(t, "x", ExprToString(&ast.ParenExpr{X: &ast.Ident{Name: "x"}}))
	assert.Equal(t, "x", ExprToString(&ast.StarExpr{X: &ast.Ident{Name: "x"}}))
	assert.Equal(t, "", ExprToString(&ast.BasicLit{Kind: token.INT, Value: "1"}))
}

func TestIsLogStringCall(t *testing.T) {
	call := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "log"},
			Sel: &ast.Ident{Name: "String"},
		},
	}
	assert.True(t, IsLogStringCall(call))

	call.Fun = &ast.SelectorExpr{X: &ast.Ident{Name: "log"}, Sel: &ast.Ident{Name: "Int"}}
	assert.False(t, IsLogStringCall(call))

	assert.False(t, IsLogStringCall(&ast.CallExpr{Fun: &ast.Ident{Name: "String"}}))
}

func TestIsFmtSprintfCall(t *testing.T) {
	call := &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "fmt"},
			Sel: &ast.Ident{Name: "Sprintf"},
		},
	}
	assert.True(t, IsFmtSprintfCall(call))

	call.Fun = &ast.SelectorExpr{X: &ast.Ident{Name: "fmt"}, Sel: &ast.Ident{Name: "Println"}}
	assert.False(t, IsFmtSprintfCall(call))

	assert.False(t, IsFmtSprintfCall(&ast.CallExpr{Fun: &ast.Ident{Name: "Sprintf"}}))
}

func TestIsStringMethodCall(t *testing.T) {
	call := &ast.CallExpr{
		Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "x"}, Sel: &ast.Ident{Name: "String"}},
		Args: nil,
	}
	assert.True(t, IsStringMethodCall(call))

	call.Args = []ast.Expr{&ast.Ident{Name: "y"}}
	assert.False(t, IsStringMethodCall(call))

	assert.False(t, IsStringMethodCall(&ast.CallExpr{Fun: &ast.Ident{Name: "String"}}))
}
