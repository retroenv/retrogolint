// Package loggingperformance contains rules that flag inefficient logging patterns.
package loggingperformance

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingEfficiencyRule detects inefficient logging patterns with method calls.
type LoggingEfficiencyRule struct{}

// NewLoggingEfficiencyRule creates a new LoggingEfficiencyRule.
func NewLoggingEfficiencyRule() *LoggingEfficiencyRule {
	return &LoggingEfficiencyRule{}
}

// Name returns the rule name.
func (r *LoggingEfficiencyRule) Name() string {
	return "logging-efficiency"
}

// Description returns the rule description.
func (r *LoggingEfficiencyRule) Description() string {
	return "Use lazy evaluation for method calls in log fields to avoid execution when logging is disabled"
}

// Severity returns the default severity.
func (r *LoggingEfficiencyRule) Severity() violation.Severity {
	return violation.SeverityInfo
}

// Category returns the rule category.
func (r *LoggingEfficiencyRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingEfficiencyRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		if !api.IsLoggerMethod(call) {
			return true
		}

		for i := 1; i < len(call.Args); i++ {
			fieldCall, ok := call.Args[i].(*ast.CallExpr)
			if !ok {
				continue
			}

			if !api.IsLogFieldCall(fieldCall) {
				continue
			}

			if len(fieldCall.Args) < 2 {
				continue
			}

			valueExpr := fieldCall.Args[1]

			if _, isCall := valueExpr.(*ast.CallExpr); isCall {
				if r.isLazyEvaluation(fieldCall) {
					continue
				}

				pos := fset.Position(fieldCall.Pos())

				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "Use lazy evaluation for method calls in log fields (e.g., log.StringFunc)",
					Position: pos,
					Severity: r.Severity(),
				})
			}
		}

		return true
	})

	return violations
}

// isLazyEvaluation checks if the field call is already using lazy evaluation.
func (r *LoggingEfficiencyRule) isLazyEvaluation(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	switch sel.Sel.Name {
	case "StringFunc", "IntFunc", "Int64Func", "UintFunc", "Uint64Func", "Float64Func":
		return true
	}

	return false
}
