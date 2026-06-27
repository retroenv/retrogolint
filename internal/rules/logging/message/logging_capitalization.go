// Package loggingmessage contains rules that inspect log message content and nil checks.
package loggingmessage

import (
	"go/ast"
	"go/token"
	"unicode"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingCapitalizationRule checks that log messages start with an uppercase letter.
type LoggingCapitalizationRule struct{}

// NewLoggingCapitalizationRule creates a new LoggingCapitalizationRule.
func NewLoggingCapitalizationRule() *LoggingCapitalizationRule {
	return &LoggingCapitalizationRule{}
}

// Name returns the rule name.
func (r *LoggingCapitalizationRule) Name() string {
	return "logging-capitalization"
}

// Description returns the rule description.
func (r *LoggingCapitalizationRule) Description() string {
	return "Log messages must start with an uppercase letter"
}

// Severity returns the default severity.
func (r *LoggingCapitalizationRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingCapitalizationRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingCapitalizationRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
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

		if logCall.MessageArgIndex == api.NoLogArgument || logCall.MessageArgIndex >= len(call.Args) {
			return true
		}

		messageArg := call.Args[logCall.MessageArgIndex]
		message, ok := api.ExtractStringLiteral(messageArg)
		if !ok {
			return true
		}

		if message == "" {
			return true
		}

		firstRune := rune(message[0])
		if !unicode.IsUpper(firstRune) && unicode.IsLetter(firstRune) {
			pos := fset.Position(messageArg.Pos())

			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  "Log message should start with an uppercase letter",
				Position: pos,
				Severity: r.Severity(),
			})
		}

		return true
	})

	return violations
}
