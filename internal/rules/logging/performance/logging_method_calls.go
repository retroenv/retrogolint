package loggingperformance

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingMethodCallsRule detects inefficient method calls in log fields.
type LoggingMethodCallsRule struct{}

// NewLoggingMethodCallsRule creates a new LoggingMethodCallsRule.
func NewLoggingMethodCallsRule() *LoggingMethodCallsRule {
	return &LoggingMethodCallsRule{}
}

// Name returns the rule name.
func (r *LoggingMethodCallsRule) Name() string {
	return "logging-method-calls"
}

// Description returns the rule description.
func (r *LoggingMethodCallsRule) Description() string {
	return "Use log.Stringer() instead of log.String() with .String() method call"
}

// Severity returns the default severity.
func (r *LoggingMethodCallsRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingMethodCallsRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingMethodCallsRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		logCall, ok := api.GetLogCallInfo(call)
		if !ok {
			return true
		}

		for i := logCall.FieldStartIndex; i < len(call.Args); i++ {
			fieldCall, ok := call.Args[i].(*ast.CallExpr)
			if !ok {
				continue
			}

			if !api.IsLogStringCall(fieldCall) {
				continue
			}

			if len(fieldCall.Args) < 2 {
				continue
			}

			methodCall, ok := fieldCall.Args[1].(*ast.CallExpr)
			if !ok {
				continue
			}

			if api.IsStringMethodCall(methodCall) {
				pos := fset.Position(fieldCall.Pos())

				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "Use log.Stringer() instead of log.String() with .String() method",
					Position: pos,
					Severity: r.Severity(),
				})
			}
		}

		return true
	})

	return violations
}
