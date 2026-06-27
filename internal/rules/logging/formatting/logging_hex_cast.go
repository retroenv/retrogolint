// Package loggingformatting contains rules for log formatting, preferring dedicated helpers over fmt.
package loggingformatting

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingHexCastRule detects unnecessary widening casts before log.Hex fields.
type LoggingHexCastRule struct{}

// NewLoggingHexCastRule creates a new LoggingHexCastRule.
func NewLoggingHexCastRule() *LoggingHexCastRule {
	return &LoggingHexCastRule{}
}

// Name returns the rule name.
func (r *LoggingHexCastRule) Name() string {
	return "logging-hex-cast"
}

// Description returns the rule description.
func (r *LoggingHexCastRule) Description() string {
	return "Avoid widening casts such as uint64(uint16(...)) before log.Hex to keep the original value width"
}

// Severity returns the default severity.
func (r *LoggingHexCastRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingHexCastRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingHexCastRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
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
			if !ok || !isLogHexCall(fieldCall) {
				continue
			}

			if r.hasUnnecessaryWidening(fieldCall) {
				pos := fset.Position(fieldCall.Pos())
				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "Avoid widening casts (e.g., uint64(uint16(...))) before log.Hex - pass the original unsigned value",
					Position: pos,
					Severity: r.Severity(),
				})
			}
		}

		return true
	})

	return violations
}

func (r *LoggingHexCastRule) hasUnnecessaryWidening(call *ast.CallExpr) bool {
	if len(call.Args) < 2 {
		return false
	}

	val := unwrapExpr(call.Args[1])
	target, ok := val.(*ast.CallExpr)
	if !ok || !isUint64Conversion(target) || len(target.Args) == 0 {
		return false
	}

	inner := unwrapExpr(target.Args[0])
	innerCall, ok := inner.(*ast.CallExpr)
	if !ok {
		return false
	}

	return isUnsignedConversion(innerCall)
}

func isUint64Conversion(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "uint64"
}

func isUnsignedConversion(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	switch ident.Name {
	case "uint8", "uint16", "uint32", "uint":
		return len(call.Args) == 1
	default:
		return false
	}
}

func unwrapExpr(expr ast.Expr) ast.Expr {
	for {
		switch e := expr.(type) {
		case *ast.ParenExpr:
			expr = e.X
			continue
		case *ast.StarExpr:
			expr = e.X
			continue
		}
		break
	}

	return expr
}

func isLogHexCall(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Hex" {
		return false
	}

	if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "log" {
		return true
	}

	return false
}
