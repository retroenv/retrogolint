// Package api exposes shared types and helpers used by the rules packages.
package api

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"github.com/retroenv/retrogolint/internal/violation"
)

// Rule category and log field type constants used across rule implementations.
const (
	CategoryLogging     = "logging"
	CategoryCollections = "collections"
	CategoryCodeQuality = "codequality"
	LogFieldTypeString  = "String"
	PackageFmt          = "fmt"
)

// Rule represents a linting rule that can detect violations.
type Rule interface {
	Name() string
	Description() string
	Severity() violation.Severity
	Category() string
	Check(fset *token.FileSet, file *ast.File) []violation.Violation
}

// IsLoggerMethod checks if a call expression is a logger method call.
func IsLoggerMethod(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if ident, ok := sel.X.(*ast.Ident); ok {
		switch ident.Name {
		case "fmt", "errors", "log", "t", "b":
			return false
		}
	}

	switch sel.Sel.Name {
	case "Debug", "Info", "Warn", "Error", "Fatal", "Panic",
		"Debugf", "Infof", "Warnf", "Errorf", "Fatalf", "Panicf":
		return true
	}

	return false
}

// IsLogFieldCall checks if a call expression is a log field constructor.
func IsLogFieldCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	switch sel.Sel.Name {
	case LogFieldTypeString, "Int", "Int64", "Uint", "Uint64", "Float64", "Bool",
		"Duration", "Time", "Error", "Stringer", "Hex", "Type",
		"StringFunc", "IntFunc":
		return true
	}
	return false
}

// ExtractStringLiteral extracts a string literal value from an expression.
func ExtractStringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return value, true
}

// IsSnakeCase checks if a string uses snake_case format.
func IsSnakeCase(s string) bool {
	if s == "" {
		return true
	}

	if s[0] != '_' && (s[0] < 'a' || s[0] > 'z') {
		return false
	}

	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '_' {
			return false
		}
	}

	return true
}

// ToSnakeCase converts a string to snake_case.
func ToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	result.Grow(len(s) + 5)

	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// IsLogStringCall checks if a call is to log.String().
func IsLogStringCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == LogFieldTypeString
}

// IsFmtSprintfCall checks if a call is to fmt.Sprintf().
func IsFmtSprintfCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Sprintf" {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == PackageFmt
}

// IsStringMethodCall checks if a call is to .String() method with no arguments.
func IsStringMethodCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	return sel.Sel.Name == LogFieldTypeString && len(call.Args) == 0
}

// ContainsFormatVerb checks if a format string contains any of the specified format verbs.
func ContainsFormatVerb(format string, verbs ...rune) bool {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		if i+1 < len(format) && format[i+1] == '%' {
			i++
			continue
		}

		for j := i + 1; j < len(format); j++ {
			c := rune(format[j])
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				for _, v := range verbs {
					if c == v {
						return true
					}
				}
				break
			}
		}
	}
	return false
}

// ExprToString converts an AST expression into a string representation.
func ExprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		if left := ExprToString(e.X); left != "" {
			return left + "." + e.Sel.Name
		}
	case *ast.ParenExpr:
		return ExprToString(e.X)
	case *ast.StarExpr:
		return ExprToString(e.X)
	}
	return ""
}
