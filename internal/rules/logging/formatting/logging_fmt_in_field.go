package loggingformatting

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const (
	// fmt function names.
	fmtSprintf  = "Sprintf"
	fmtSprint   = "Sprint"
	fmtSprintln = "Sprintln"
	fmtErrorf   = "Errorf"

	// Format verbs.
	formatVerbGeneric = "%v"
	formatVerbString  = "%s"
	formatVerbDecimal = "%d"

	// Suggestion messages.
	suggestionLogAny            = "use log.Any() instead"
	suggestionLogAnyOrDirect    = "use log.Any() or the value directly if it's already a string"
	suggestionStringDirect      = "use the string value directly or log.Stringer() for fmt.Stringer types"
	suggestionLogInt            = "use log.Int() or appropriate integer log field"
	suggestionLogHex            = "use log.Hex() for hex formatting"
	suggestionLogType           = "use log.Type() for type information"
	suggestionLogUintptr        = "use log.Uintptr() for pointer values"
	suggestionLogFloat          = "use log.Float64() or log.Float32()"
	suggestionLogBool           = "use log.Bool()"
	suggestionLogError          = "use log.Error() field or wrap the error separately"
	suggestionStructuredFields  = "consider using separate structured log fields instead"
	suggestionLogAnyOrTypeField = "use log.Any() or a type-specific log field"
)

// LoggingFmtInFieldRule detects fmt package function calls in log field values.
type LoggingFmtInFieldRule struct{}

// NewLoggingFmtInFieldRule creates a new LoggingFmtInFieldRule.
func NewLoggingFmtInFieldRule() *LoggingFmtInFieldRule {
	return &LoggingFmtInFieldRule{}
}

// Name returns the rule name.
func (r *LoggingFmtInFieldRule) Name() string {
	return "logging-fmt-in-field"
}

// Description returns the rule description.
func (r *LoggingFmtInFieldRule) Description() string {
	return "Avoid using fmt package functions in log field values; use log.Any() or type-specific log fields instead"
}

// Severity returns the default severity.
func (r *LoggingFmtInFieldRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingFmtInFieldRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingFmtInFieldRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
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

			if !api.IsLogFieldCall(fieldCall) {
				continue
			}

			for argIdx, arg := range fieldCall.Args {
				if argIdx == 0 {
					continue
				}

				fmtCall, fmtFunc := r.extractFmtCall(arg)
				if fmtCall == nil {
					continue
				}

				pos := fset.Position(fmtCall.Pos())
				logFieldType := r.getLogFieldType(fieldCall)
				message := r.buildMessage(fmtFunc, logFieldType, fmtCall)

				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  message,
					Position: pos,
					Severity: r.Severity(),
				})
			}
		}

		return true
	})

	return violations
}

// extractFmtCall checks if an expression is an fmt.* function call and returns it.
func (r *LoggingFmtInFieldRule) extractFmtCall(expr ast.Expr) (*ast.CallExpr, string) {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, ""
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, ""
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok || ident.Name != api.PackageFmt {
		return nil, ""
	}

	funcName := sel.Sel.Name
	switch funcName {
	case fmtSprintf, fmtSprint, fmtSprintln, fmtErrorf:
		return call, funcName
	}

	return nil, ""
}

// getLogFieldType extracts the log field type name.
func (r *LoggingFmtInFieldRule) getLogFieldType(call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "field"
	}
	return sel.Sel.Name
}

// buildMessage creates a helpful violation message based on the fmt function used.
func (r *LoggingFmtInFieldRule) buildMessage(fmtFunc, logFieldType string, fmtCall *ast.CallExpr) string {
	suggestion := r.getSuggestion(fmtFunc, fmtCall)

	if suggestion != "" {
		return fmt.Sprintf("Avoid %s in log.%s(); %s", fmtFunc, logFieldType, suggestion)
	}

	return fmt.Sprintf("Avoid %s in log.%s(); %s", fmtFunc, logFieldType, suggestionLogAnyOrTypeField)
}

// getSuggestion provides a specific suggestion based on the fmt function and its usage.
func (r *LoggingFmtInFieldRule) getSuggestion(fmtFunc string, fmtCall *ast.CallExpr) string {
	switch fmtFunc {
	case fmtSprintf:
		return r.getSprintfSuggestion(fmtCall)
	case fmtSprint, fmtSprintln:
		return suggestionLogAnyOrDirect
	case fmtErrorf:
		return suggestionLogError
	}
	return ""
}

// getSprintfSuggestion analyzes fmt.Sprintf usage and suggests better alternatives.
func (r *LoggingFmtInFieldRule) getSprintfSuggestion(call *ast.CallExpr) string {
	if len(call.Args) < 1 {
		return ""
	}

	formatStr, ok := api.ExtractStringLiteral(call.Args[0])
	if !ok {
		return ""
	}

	if formatStr == formatVerbGeneric {
		return suggestionLogAny
	}

	if formatStr == formatVerbString {
		return suggestionStringDirect
	}

	if formatStr == formatVerbDecimal {
		return suggestionLogInt
	}

	if api.ContainsFormatVerb(formatStr, 'x', 'X') {
		return suggestionLogHex
	}

	if api.ContainsFormatVerb(formatStr, 'T') {
		return suggestionLogType
	}

	if api.ContainsFormatVerb(formatStr, 'p') {
		return suggestionLogUintptr
	}

	if api.ContainsFormatVerb(formatStr, 'f', 'g', 'e', 'E', 'G') {
		return suggestionLogFloat
	}

	if api.ContainsFormatVerb(formatStr, 't') {
		return suggestionLogBool
	}

	verbCount := strings.Count(formatStr, "%") - strings.Count(formatStr, "%%")*2
	if verbCount > 1 {
		return suggestionStructuredFields
	}

	return ""
}
