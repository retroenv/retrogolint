// Package loggingfields contains rules that validate logger field usage.
package loggingfields

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingFieldCasingRule checks that logger field names use snake_case.
type LoggingFieldCasingRule struct{}

// NewLoggingFieldCasingRule creates a new LoggingFieldCasingRule.
func NewLoggingFieldCasingRule() *LoggingFieldCasingRule {
	return &LoggingFieldCasingRule{}
}

// Name returns the rule name.
func (r *LoggingFieldCasingRule) Name() string {
	return "logging-field-casing"
}

// Description returns the rule description.
func (r *LoggingFieldCasingRule) Description() string {
	return "Logger field names should use snake_case convention"
}

// Severity returns the default severity.
func (r *LoggingFieldCasingRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingFieldCasingRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingFieldCasingRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
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

			if len(fieldCall.Args) == 0 {
				continue
			}

			key, ok := api.ExtractStringLiteral(fieldCall.Args[0])
			if !ok {
				continue
			}

			if !api.IsSnakeCase(key) {
				pos := fset.Position(fieldCall.Args[0].Pos())

				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "Logger field name should use snake_case",
					Position: pos,
					Severity: r.Severity(),
				})
			}
		}

		return true
	})

	return violations
}
