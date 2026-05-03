// Package collections contains rules for collection usage patterns.
package collections

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// MapSetRule detects maps that are being used as sets and suggests retrogolib set.Set.
type MapSetRule struct{}

// NewMapSetRule creates a new MapSetRule.
func NewMapSetRule() *MapSetRule {
	return &MapSetRule{}
}

// Name returns the rule name.
func (r *MapSetRule) Name() string {
	return "collections-map-set"
}

// Description returns the rule description.
func (r *MapSetRule) Description() string {
	return "Use retrogolib set.Set instead of map[K]bool or map[K]struct{} when modeling sets"
}

// Severity returns the default severity.
func (r *MapSetRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *MapSetRule) Category() string {
	return api.CategoryCollections
}

// Check analyzes a file for violations.
func (r *MapSetRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		mapType, ok := n.(*ast.MapType)
		if !ok {
			return true
		}

		if !isSetValueType(mapType.Value) {
			return true
		}

		pos := fset.Position(mapType.Pos())

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "Use retrogolib set.Set instead of map[K]bool or map[K]struct{}",
			Position: pos,
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}

func isSetValueType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "bool" {
		return true
	}

	if structType, ok := expr.(*ast.StructType); ok {
		return structType.Fields == nil || len(structType.Fields.List) == 0
	}

	return false
}
