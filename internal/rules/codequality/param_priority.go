// Package codequality contains rules for code quality and organization.
package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// ParamPriorityRule checks that function parameters follow the correct ordering:
// context.Context first, then logger, then everything else.
type ParamPriorityRule struct{}

// NewParamPriorityRule creates a new ParamPriorityRule.
func NewParamPriorityRule() *ParamPriorityRule {
	return &ParamPriorityRule{}
}

// Name returns the rule name.
func (r *ParamPriorityRule) Name() string {
	return "codequality-param-priority"
}

// Description returns the rule description.
func (r *ParamPriorityRule) Description() string {
	return "Function parameters should be ordered: context first, then logger, then other parameters"
}

// Severity returns the default severity.
func (r *ParamPriorityRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *ParamPriorityRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for violations.
func (r *ParamPriorityRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) < 2 {
			return true
		}

		violations = r.checkParams(fset, funcDecl, violations)
		return true
	})

	return violations
}

// checkParams checks the parameter ordering of a function declaration.
func (r *ParamPriorityRule) checkParams(fset *token.FileSet, funcDecl *ast.FuncDecl, violations []violation.Violation) []violation.Violation {
	params := funcDecl.Type.Params.List

	// Track position categories for each parameter.
	// Category 0 = context, 1 = logger, 2 = other.
	paramInfos := make([]paramCategory, 0, len(params))
	for _, field := range params {
		paramInfos = append(paramInfos, classifyParam(field))
	}

	for i := 1; i < len(paramInfos); i++ {
		if paramInfos[i] < paramInfos[i-1] {
			pos := fset.Position(params[i].Pos())
			msg := r.formatMessage(funcDecl.Name.Name, paramInfos[i], paramInfos[i-1])
			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  msg,
				Position: pos,
				Severity: r.Severity(),
			})
			break
		}
	}

	return violations
}

// formatMessage creates a descriptive message for the violation.
func (r *ParamPriorityRule) formatMessage(funcName string, current, previous paramCategory) string {
	return funcName + ": " + current.String() + " parameter should come before " + previous.String() + " parameter"
}

// paramCategory represents the priority category of a function parameter.
type paramCategory int

const (
	paramCategoryContext paramCategory = iota
	paramCategoryLogger
	paramCategoryOther
)

// String returns a human-readable name for the parameter category.
func (c paramCategory) String() string {
	switch c {
	case paramCategoryContext:
		return "context"
	case paramCategoryLogger:
		return "logger"
	case paramCategoryOther:
		return "other"
	default:
		return "unknown"
	}
}

// classifyParam determines the priority category of a function parameter.
func classifyParam(field *ast.Field) paramCategory {
	typeName := resolveTypeName(field.Type)

	if typeName == "context.Context" || typeName == "Context" {
		return paramCategoryContext
	}

	if isLoggerType(typeName) {
		return paramCategoryLogger
	}

	return paramCategoryOther
}

// resolveTypeName extracts the type name from an AST expression, handling
// pointers, selectors, and identifiers.
func resolveTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return resolveTypeName(t.X)
	}
	return ""
}

// isLoggerType checks if a type name represents a known logger type.
func isLoggerType(typeName string) bool {
	switch typeName {
	case "slog.Logger",
		"slog.Handler",
		"log.Logger",
		"zap.Logger",
		"zap.SugaredLogger",
		"zerolog.Logger",
		"logrus.Logger",
		"logrus.Entry",
		"Logger":
		return true
	}
	return false
}
