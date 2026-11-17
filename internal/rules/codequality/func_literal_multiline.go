package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// FuncLiteralMultilineRule detects function literals that keep a non-empty body on one line.
type FuncLiteralMultilineRule struct{}

// NewFuncLiteralMultilineRule creates a new FuncLiteralMultilineRule.
func NewFuncLiteralMultilineRule() *FuncLiteralMultilineRule {
	return &FuncLiteralMultilineRule{}
}

// Name returns the rule name.
func (r *FuncLiteralMultilineRule) Name() string {
	return "codequality-func-literal-multiline"
}

// Description returns the rule description.
func (r *FuncLiteralMultilineRule) Description() string {
	return "Function literals passed as call arguments or assigned values should use multiple lines"
}

// Severity returns the default severity.
func (r *FuncLiteralMultilineRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *FuncLiteralMultilineRule) Category() string {
	return api.CategoryCodeQuality
}

// Check reports function literals that keep a non-empty body on one line.
func (r *FuncLiteralMultilineRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	check := func(literal *ast.FuncLit) {
		if len(literal.Body.List) == 0 {
			return
		}

		if fset.Position(literal.Body.Lbrace).Line != fset.Position(literal.Body.Rbrace).Line {
			return
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "function literal should use multiple lines",
			Position: fset.Position(literal.Pos()),
			Severity: r.Severity(),
		})
	}

	ast.Inspect(file, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.CallExpr:
			for _, arg := range node.Args {
				if literal, ok := ast.Unparen(arg).(*ast.FuncLit); ok {
					check(literal)
				}
			}

		case *ast.AssignStmt:
			for _, value := range node.Rhs {
				if literal, ok := ast.Unparen(value).(*ast.FuncLit); ok {
					check(literal)
				}
			}

		case *ast.ValueSpec:
			for _, value := range node.Values {
				if literal, ok := ast.Unparen(value).(*ast.FuncLit); ok {
					check(literal)
				}
			}
		}

		return true
	})

	return violations
}
