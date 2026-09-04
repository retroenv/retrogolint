package collections

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// SetSortRule detects manual collection and sorting of set values.
type SetSortRule struct{}

// NewSetSortRule creates a new SetSortRule.
func NewSetSortRule() *SetSortRule {
	return &SetSortRule{}
}

// Name returns the rule name.
func (r *SetSortRule) Name() string {
	return "collections-set-sort"
}

// Description returns the rule description.
func (r *SetSortRule) Description() string {
	return "Use set.Sorted instead of manually collecting and sorting set values"
}

// Severity returns the default severity.
func (r *SetSortRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *SetSortRule) Category() string {
	return api.CategoryCollections
}

// Check analyzes a file for manual sorted set projections.
func (r *SetSortRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	if !importsRetrogolibSet(file) {
		return nil
	}

	var violations []violation.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		block, ok := n.(*ast.BlockStmt)
		if !ok {
			return true
		}

		for index := range len(block.List) - 1 {
			rangeStmt, ok := block.List[index].(*ast.RangeStmt)
			if !ok || !isSetCollectionLoop(rangeStmt) {
				continue
			}

			sliceName, ok := appendedSliceName(rangeStmt)
			if !ok || !sortsSlice(block.List[index+1], sliceName) {
				continue
			}

			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  "Use set.Sorted instead of manually collecting and sorting set values",
				Position: fset.Position(rangeStmt.Pos()),
				Severity: r.Severity(),
			})
		}

		return true
	})

	return violations
}

func isSetCollectionLoop(rangeStmt *ast.RangeStmt) bool {
	_, isBlank := rangeStmt.Key.(*ast.Ident)
	return rangeStmt.Value == nil && isBlank
}

func appendedSliceName(rangeStmt *ast.RangeStmt) (string, bool) {
	keyName, ok := rangeKeyName(rangeStmt)
	if !ok {
		return "", false
	}

	assignment, ok := singleAssignment(rangeStmt.Body)
	if !ok {
		return "", false
	}

	return appendAssignmentSliceName(assignment, keyName)
}

func rangeKeyName(rangeStmt *ast.RangeStmt) (string, bool) {
	key, ok := rangeStmt.Key.(*ast.Ident)
	if !ok || key.Name == "_" || len(rangeStmt.Body.List) != 1 {
		return "", false
	}

	return key.Name, true
}

func singleAssignment(body *ast.BlockStmt) (*ast.AssignStmt, bool) {
	assignment, ok := body.List[0].(*ast.AssignStmt)
	if !ok || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
		return nil, false
	}

	return assignment, true
}

func appendAssignmentSliceName(assignment *ast.AssignStmt, keyName string) (string, bool) {
	slice, ok := assignment.Lhs[0].(*ast.Ident)
	if !ok {
		return "", false
	}

	call, ok := assignment.Rhs[0].(*ast.CallExpr)
	if !ok || len(call.Args) != 2 || !isAppendCall(call) {
		return "", false
	}

	appendedSlice, sliceOK := call.Args[0].(*ast.Ident)
	appendedValue, valueOK := call.Args[1].(*ast.Ident)
	if !sliceOK || !valueOK || slice.Name != appendedSlice.Name || keyName != appendedValue.Name {
		return "", false
	}

	return slice.Name, true
}

func isAppendCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "append"
}

func sortsSlice(statement ast.Stmt, sliceName string) bool {
	expression, ok := statement.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := expression.X.(*ast.CallExpr)
	if !ok || len(call.Args) != 1 {
		return false
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Sort" {
		return false
	}

	ident, ok := call.Args[0].(*ast.Ident)
	return ok && ident.Name == sliceName
}
