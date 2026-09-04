package collections

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// SetMembershipRule detects membership checks that scan a set.
type SetMembershipRule struct{}

// NewSetMembershipRule creates a new SetMembershipRule.
func NewSetMembershipRule() *SetMembershipRule {
	return &SetMembershipRule{}
}

// Name returns the rule name.
func (r *SetMembershipRule) Name() string {
	return "collections-set-membership"
}

// Description returns the rule description.
func (r *SetMembershipRule) Description() string {
	return "Use Set.Contains instead of scanning a set for one value"
}

// Severity returns the default severity.
func (r *SetMembershipRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *SetMembershipRule) Category() string {
	return api.CategoryCollections
}

// Check analyzes a file for set membership scans.
func (r *SetMembershipRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	if !importsRetrogolibSet(file) {
		return nil
	}

	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		rangeStmt, ok := n.(*ast.RangeStmt)
		if !ok || !isMembershipScan(rangeStmt) {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "Use Set.Contains instead of scanning a set for one value",
			Position: fset.Position(rangeStmt.Pos()),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}

func isMembershipScan(rangeStmt *ast.RangeStmt) bool {
	key, ok := rangeStmt.Key.(*ast.Ident)
	if !ok || key.Name == "_" || rangeStmt.Value != nil || len(rangeStmt.Body.List) != 1 {
		return false
	}

	ifStmt, ok := rangeStmt.Body.List[0].(*ast.IfStmt)
	if !ok || ifStmt.Else != nil || len(ifStmt.Body.List) != 1 {
		return false
	}

	return comparesIdentToValue(ifStmt.Cond, key.Name) && isReturningStatement(ifStmt.Body.List[0])
}

func comparesIdentToValue(expression ast.Expr, name string) bool {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok || binary.Op != token.EQL {
		return false
	}

	left, leftOK := binary.X.(*ast.Ident)
	right, rightOK := binary.Y.(*ast.Ident)
	return (leftOK && left.Name == name && (!rightOK || right.Name != name)) ||
		(rightOK && right.Name == name && (!leftOK || left.Name != name))
}

func isReturningStatement(statement ast.Stmt) bool {
	_, ok := statement.(*ast.ReturnStmt)
	return ok
}
