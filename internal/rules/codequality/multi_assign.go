package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// MultiAssignRule detects assignments with multiple explicit values.
type MultiAssignRule struct{}

// NewMultiAssignRule creates a new MultiAssignRule.
func NewMultiAssignRule() *MultiAssignRule {
	return &MultiAssignRule{}
}

// Name returns the rule name.
func (r *MultiAssignRule) Name() string {
	return "codequality-multi-assign"
}

// Description returns the rule description.
func (r *MultiAssignRule) Description() string {
	return "Avoid assignments with multiple explicit values"
}

// Severity returns the default severity.
func (r *MultiAssignRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *MultiAssignRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes assignment statements.
func (r *MultiAssignRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(node ast.Node) bool {
		assignment, ok := node.(*ast.AssignStmt)
		if !ok || len(assignment.Rhs) < 2 {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "avoid assigning multiple explicit values in one statement",
			Position: fset.Position(assignment.Pos()),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}
