package loggingformatting

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingHexFormattingRule detects hex formatting patterns that should use log.Hex().
type LoggingHexFormattingRule struct{}

// NewLoggingHexFormattingRule creates a new LoggingHexFormattingRule.
func NewLoggingHexFormattingRule() *LoggingHexFormattingRule {
	return &LoggingHexFormattingRule{}
}

// Name returns the rule name.
func (r *LoggingHexFormattingRule) Name() string {
	return "logging-hex-formatting"
}

// Description returns the rule description.
func (r *LoggingHexFormattingRule) Description() string {
	return "Use log.Hex() instead of fmt.Sprintf with hex formatting for better performance"
}

// Severity returns the default severity.
func (r *LoggingHexFormattingRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingHexFormattingRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingHexFormattingRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
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

			if !api.IsLogStringCall(fieldCall) {
				continue
			}

			if len(fieldCall.Args) < 2 {
				continue
			}

			sprintfCall, ok := fieldCall.Args[1].(*ast.CallExpr)
			if !ok {
				continue
			}

			if !api.IsFmtSprintfCall(sprintfCall) {
				continue
			}

			if len(sprintfCall.Args) < 1 {
				continue
			}

			formatStr, ok := api.ExtractStringLiteral(sprintfCall.Args[0])
			if !ok {
				continue
			}

			if api.ContainsFormatVerb(formatStr, 'x', 'X') {
				pos := fset.Position(fieldCall.Pos())

				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "Use log.Hex() instead of fmt.Sprintf with hex formatting",
					Position: pos,
					Severity: r.Severity(),
				})
			}
		}

		return true
	})

	return violations
}
