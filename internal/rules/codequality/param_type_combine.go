package codequality

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// ParamTypeCombineRule detects consecutive parameters that repeat the same type.
type ParamTypeCombineRule struct{}

// NewParamTypeCombineRule creates a new ParamTypeCombineRule.
func NewParamTypeCombineRule() *ParamTypeCombineRule {
	return &ParamTypeCombineRule{}
}

// Name returns the rule name.
func (r *ParamTypeCombineRule) Name() string {
	return "codequality-param-type-combine"
}

// Description returns the rule description.
func (r *ParamTypeCombineRule) Description() string {
	return "Combine consecutive function parameters that have the same type"
}

// Severity returns the default severity.
func (r *ParamTypeCombineRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *ParamTypeCombineRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes function and method parameters.
func (r *ParamTypeCombineRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Type.Params == nil || hasSignatureComments(file, function) {
			continue
		}

		violations = r.checkFields(fset, function, violations)
	}

	return violations
}

func (r *ParamTypeCombineRule) checkFields(fset *token.FileSet, function *ast.FuncDecl,
	violations []violation.Violation) []violation.Violation {

	fields := function.Type.Params.List
	for index := 1; index < len(fields); index++ {
		previous := fields[index-1]
		current := fields[index]
		if !fieldsCanCombine(previous, current) {
			continue
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  function.Name.Name + ": consecutive parameters with type " + types.ExprString(current.Type) + " should share one type declaration",
			Position: fset.Position(current.Pos()),
			Severity: r.Severity(),
		})

		for index+1 < len(fields) && fieldsCanCombine(fields[index], fields[index+1]) {
			index++
		}
	}

	return violations
}

func fieldsCanCombine(left, right *ast.Field) bool {
	return len(left.Names) > 0 && len(right.Names) > 0 &&
		types.ExprString(left.Type) == types.ExprString(right.Type)
}
