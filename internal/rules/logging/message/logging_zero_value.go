package loggingmessage

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingZeroValueRule detects usage of &log.Logger{} which creates an unsafe zero-value logger.
type LoggingZeroValueRule struct{}

// NewLoggingZeroValueRule creates a new LoggingZeroValueRule.
func NewLoggingZeroValueRule() *LoggingZeroValueRule {
	return &LoggingZeroValueRule{}
}

// Name returns the rule name.
func (r *LoggingZeroValueRule) Name() string {
	return "logging-zero-value"
}

// Description returns the rule description.
func (r *LoggingZeroValueRule) Description() string {
	return "Use of &log.Logger{} creates an unsafe zero-value logger that will crash - use nil instead"
}

// Severity returns the default severity.
func (r *LoggingZeroValueRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingZeroValueRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingZeroValueRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		// Look for unary expressions (& operator)
		unaryExpr, ok := n.(*ast.UnaryExpr)
		if !ok || unaryExpr.Op != token.AND {
			return true
		}

		compLit, ok := unaryExpr.X.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if it's an empty composite literal (no elements)
		if len(compLit.Elts) > 0 {
			return true
		}

		if !r.isLogLogger(compLit.Type) {
			return true
		}

		pos := fset.Position(unaryExpr.Pos())

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "Use of &log.Logger{} creates unsafe zero-value logger - use nil instead (logger methods are nil-safe)",
			Position: pos,
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}

// isLogLogger checks if a type expression is log.Logger.
func (r *LoggingZeroValueRule) isLogLogger(typ ast.Expr) bool {
	sel, ok := typ.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if sel.Sel.Name != "Logger" {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == "log"
}
