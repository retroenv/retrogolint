package collections

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const retrogolibSetImportPath = "github.com/retroenv/retrogolib/set"

// SetProjectionRule detects the legacy Set.ToSlice projection.
type SetProjectionRule struct{}

// NewSetProjectionRule creates a new SetProjectionRule.
func NewSetProjectionRule() *SetProjectionRule {
	return &SetProjectionRule{}
}

// Name returns the rule name.
func (r *SetProjectionRule) Name() string {
	return "collections-set-projection"
}

// Description returns the rule description.
func (r *SetProjectionRule) Description() string {
	return "Use set.Sorted or set.SortedFunc instead of Set.ToSlice"
}

// Severity returns the default severity.
func (r *SetProjectionRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *SetProjectionRule) Category() string {
	return api.CategoryCollections
}

// Check analyzes a file for legacy set projections.
func (r *SetProjectionRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	if !importsRetrogolibSet(file) {
		return nil
	}

	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 0 || !isToSliceCall(call) {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "Use set.Sorted or set.SortedFunc instead of Set.ToSlice",
			Position: fset.Position(call.Pos()),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}

func importsRetrogolibSet(file *ast.File) bool {
	for _, importSpec := range file.Imports {
		path, err := strconv.Unquote(importSpec.Path.Value)
		if err == nil && path == retrogolibSetImportPath {
			return true
		}
	}

	return false
}

func isToSliceCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "ToSlice"
}
