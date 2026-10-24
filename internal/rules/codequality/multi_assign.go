package codequality

import (
	"go/ast"
	"go/token"
	"go/types"

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
		if !ok || len(assignment.Rhs) < 2 || isPermutation(assignment) {
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

// isPermutation reports whether the values on the right side are the targets
// on the left side in another order, such as a swap. Calls and channel
// receives can change the result when the order changes, so they are not
// allowed in a permutation.
func isPermutation(assignment *ast.AssignStmt) bool {
	if assignment.Tok != token.ASSIGN || len(assignment.Lhs) != len(assignment.Rhs) {
		return false
	}

	counts := make(map[string]int, len(assignment.Lhs))
	for _, target := range assignment.Lhs {
		if hasCallOrReceive(target) {
			return false
		}

		counts[types.ExprString(target)]++
	}

	for _, value := range assignment.Rhs {
		if hasCallOrReceive(value) {
			return false
		}

		key := types.ExprString(value)
		if counts[key] == 0 {
			return false
		}

		counts[key]--
	}

	return true
}

// hasCallOrReceive reports whether an expression contains a function call or
// a channel receive.
func hasCallOrReceive(expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.CallExpr:
			found = true

		case *ast.UnaryExpr:
			if n.Op == token.ARROW {
				found = true
			}
		}

		return !found
	})

	return found
}
