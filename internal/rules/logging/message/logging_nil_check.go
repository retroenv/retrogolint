package loggingmessage

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingNilCheckRule detects unnecessary nil checks before logger method calls.
type LoggingNilCheckRule struct{}

// NewLoggingNilCheckRule creates a new LoggingNilCheckRule.
func NewLoggingNilCheckRule() *LoggingNilCheckRule {
	return &LoggingNilCheckRule{}
}

// Name returns the rule name.
func (r *LoggingNilCheckRule) Name() string {
	return "logging-nil-check"
}

// Description returns the rule description.
func (r *LoggingNilCheckRule) Description() string {
	return "Unnecessary nil check before logger method call - logger methods already handle nil receivers"
}

// Severity returns the default severity.
func (r *LoggingNilCheckRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingNilCheckRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingNilCheckRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}

		loggerExpr := r.getLoggerNilCheckExpr(ifStmt.Cond)
		if loggerExpr == "" {
			return true
		}

		if r.bodyContainsOnlyLoggerCalls(ifStmt.Body, loggerExpr) {
			pos := fset.Position(ifStmt.Pos())

			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  "Unnecessary nil check before logger call - logger methods are nil-safe",
				Position: pos,
				Severity: r.Severity(),
			})
		}

		return true
	})

	return violations
}

// getLoggerNilCheckExpr returns the logger expression string if the condition is "logger != nil" or "nil != logger", empty string otherwise.
func (r *LoggingNilCheckRule) getLoggerNilCheckExpr(cond ast.Expr) string {
	binExpr, ok := cond.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.NEQ {
		return ""
	}

	if expr := api.ExprToString(binExpr.X); expr != "" {
		if yIdent, ok := binExpr.Y.(*ast.Ident); ok && yIdent.Name == "nil" {
			return expr
		}
	}

	if expr := api.ExprToString(binExpr.Y); expr != "" {
		if xIdent, ok := binExpr.X.(*ast.Ident); ok && xIdent.Name == "nil" {
			return expr
		}
	}

	return ""
}

// bodyContainsOnlyLoggerCalls checks if the block contains only logger method calls on the specified expression string.
func (r *LoggingNilCheckRule) bodyContainsOnlyLoggerCalls(block *ast.BlockStmt, loggerExpr string) bool {
	if block == nil || len(block.List) == 0 {
		return false
	}

	for _, stmt := range block.List {
		exprStmt, ok := stmt.(*ast.ExprStmt)
		if !ok {
			return false
		}

		call, ok := exprStmt.X.(*ast.CallExpr)
		if !ok {
			return false
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}

		if api.ExprToString(sel.X) != loggerExpr {
			return false
		}

		if !api.IsLoggerMethod(call) {
			return false
		}
	}

	return true
}
