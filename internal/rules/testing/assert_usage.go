// Package testing provides linting rules for test code patterns.
package testing

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const (
	nilLiteral = "nil"
)

// AssertUsageRule detects testing patterns that should use assert functions.
type AssertUsageRule struct{}

// NewAssertUsageRule creates a new AssertUsageRule.
func NewAssertUsageRule() *AssertUsageRule {
	return &AssertUsageRule{}
}

// Name returns the rule name.
func (r *AssertUsageRule) Name() string {
	return "assert-usage"
}

// Description returns the rule description.
func (r *AssertUsageRule) Description() string {
	return "Use assert package functions instead of manual test assertions"
}

// Severity returns the default severity.
func (r *AssertUsageRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *AssertUsageRule) Category() string {
	return "testing"
}

// Check analyzes a file for violations.
func (r *AssertUsageRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	if !strings.HasSuffix(fset.Position(file.Pos()).Filename, "_test.go") {
		return violations
	}

	ast.Inspect(file, func(n ast.Node) bool {
		r.checkNode(fset, n, &violations)
		return true
	})

	return violations
}

// checkNode checks a single AST node for violations.
func (r *AssertUsageRule) checkNode(fset *token.FileSet, n ast.Node, violations *[]violation.Violation) {
	if ifStmt, ok := n.(*ast.IfStmt); ok {
		if v := r.checkIfStatementPatterns(fset, ifStmt); v != nil {
			*violations = append(*violations, *v)
		}
		return
	}

	if exprStmt, ok := n.(*ast.ExprStmt); ok {
		if call, ok := exprStmt.X.(*ast.CallExpr); ok {
			if v := r.checkAssertEqualLen(fset, call); v != nil {
				*violations = append(*violations, *v)
			}
		}
	}
}

// checkIfStatementPatterns checks an if statement for all supported patterns.
func (r *AssertUsageRule) checkIfStatementPatterns(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	// Check nil patterns first before error patterns.
	checkers := []func(*token.FileSet, *ast.IfStmt) *violation.Violation{
		r.checkNilPattern,
		r.checkNotNilPattern,
		r.checkNoErrorPattern,
		r.checkErrorPattern,
		r.checkLenPattern,
		r.checkEqualPattern,
		r.checkNotEqualPattern,
		r.checkTruePattern,
		r.checkFalsePattern,
		r.checkComparisonPattern,
	}

	for _, checker := range checkers {
		if v := checker(fset, ifStmt); v != nil {
			return v
		}
	}

	return nil
}

// hasTestFailureBody checks if an if statement has a body with a test failure as first statement.
func (r *AssertUsageRule) hasTestFailureBody(ifStmt *ast.IfStmt) bool {
	return ifStmt.Body != nil && len(ifStmt.Body.List) > 0 && r.isTestFailure(ifStmt.Body.List[0])
}

// checkNoErrorPattern detects "if err != nil { t.Fatal/t.Fatalf(...) }" patterns.
func (r *AssertUsageRule) checkNoErrorPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	errExpr := extractNilComparison(ifStmt.Cond, token.NEQ)
	if errExpr == nil || !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	errName := api.ExprToString(errExpr)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.NoError(t, %s) instead of manual error check", errName),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkErrorPattern detects "if err == nil { t.Fatal/t.Fatalf(...) }" patterns.
func (r *AssertUsageRule) checkErrorPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	errExpr := extractNilComparison(ifStmt.Cond, token.EQL)
	if errExpr == nil || !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	errName := api.ExprToString(errExpr)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.Error(t, %s) instead of manual error check", errName),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkEqualPattern detects "if actual != expected { t.Error(...) }" patterns.
func (r *AssertUsageRule) checkEqualPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.NEQ {
		return nil
	}

	if isNilIdent(binExpr.X) || isNilIdent(binExpr.Y) {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	left := r.exprToString(binExpr.X)
	right := r.exprToString(binExpr.Y)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.Equal(t, %s, %s) instead of manual equality check", right, left),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkNotEqualPattern detects "if actual == unexpected { t.Error(...) }" patterns.
func (r *AssertUsageRule) checkNotEqualPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.EQL {
		return nil
	}

	if isNilIdent(binExpr.X) || isNilIdent(binExpr.Y) {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	left := r.exprToString(binExpr.X)
	right := r.exprToString(binExpr.Y)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.NotEqual(t, %s, %s) instead of manual inequality check", right, left),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkTruePattern detects "if !condition { t.Error(...) }" patterns.
func (r *AssertUsageRule) checkTruePattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	unaryExpr, ok := ifStmt.Cond.(*ast.UnaryExpr)
	if !ok || unaryExpr.Op != token.NOT {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	condition := r.exprToString(unaryExpr.X)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.True(t, %s) instead of manual boolean check", condition),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkFalsePattern detects "if condition { t.Error(...) }" patterns where condition is a boolean.
func (r *AssertUsageRule) checkFalsePattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	if _, ok := ifStmt.Cond.(*ast.UnaryExpr); ok {
		return nil
	}
	if _, ok := ifStmt.Cond.(*ast.BinaryExpr); ok {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	condStr := r.exprToString(ifStmt.Cond)
	if condStr == "" {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.False(t, %s) instead of manual boolean check", condStr),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkNilPattern detects "if value != nil { t.Error(...) }" patterns.
func (r *AssertUsageRule) checkNilPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	valueExpr := extractNilComparison(ifStmt.Cond, token.NEQ)
	if valueExpr == nil {
		return nil
	}

	if ident, ok := valueExpr.(*ast.Ident); ok && ident.Name == "err" {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	valueName := r.exprToString(valueExpr)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.Nil(t, %s) instead of manual nil check", valueName),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkNotNilPattern detects "if value == nil { t.Fatal(...) }" patterns.
func (r *AssertUsageRule) checkNotNilPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	valueExpr := extractNilComparison(ifStmt.Cond, token.EQL)
	if valueExpr == nil {
		return nil
	}

	if ident, ok := valueExpr.(*ast.Ident); ok && ident.Name == "err" {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	valueName := r.exprToString(valueExpr)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.NotNil(t, %s) instead of manual nil check", valueName),
		Position: pos,
		Severity: r.Severity(),
	}
}

// checkLenPattern detects "if len(x) != expected { t.Error(...) }" patterns.
func (r *AssertUsageRule) checkLenPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok || binExpr.Op != token.NEQ {
		return nil
	}

	var lenCall *ast.CallExpr
	var expected ast.Expr

	if call, ok := binExpr.X.(*ast.CallExpr); ok && isLenCall(call) {
		lenCall = call
		expected = binExpr.Y
	} else if call, ok := binExpr.Y.(*ast.CallExpr); ok && isLenCall(call) {
		lenCall = call
		expected = binExpr.X
	}

	if lenCall == nil || !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	object := r.exprToString(lenCall.Args[0])
	expectedStr := r.exprToString(expected)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.Len(t, %s, %s) instead of manual length check", object, expectedStr),
		Position: pos,
		Severity: r.Severity(),
	}
}

// comparisonMapping maps a condition operator to the assert function that should be used.
var comparisonMapping = []struct {
	op         token.Token
	assertFunc string
}{
	{token.LEQ, "Greater"},
	{token.LSS, "GreaterOrEqual"},
	{token.GEQ, "Less"},
	{token.GTR, "LessOrEqual"},
}

// checkComparisonPattern detects comparison patterns like "if value <= bound { t.Error(...) }".
func (r *AssertUsageRule) checkComparisonPattern(fset *token.FileSet, ifStmt *ast.IfStmt) *violation.Violation {
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	var assertFunc string
	for _, m := range comparisonMapping {
		if binExpr.Op == m.op {
			assertFunc = m.assertFunc
			break
		}
	}
	if assertFunc == "" {
		return nil
	}

	if !r.hasTestFailureBody(ifStmt) {
		return nil
	}

	pos := fset.Position(ifStmt.Pos())
	first := r.exprToString(binExpr.X)
	second := r.exprToString(binExpr.Y)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.%s(t, %s, %s) instead of manual comparison", assertFunc, first, second),
		Position: pos,
		Severity: r.Severity(),
	}
}

// isTestFailure checks if a statement is a test failure call.
func (r *AssertUsageRule) isTestFailure(stmt ast.Stmt) bool {
	exprStmt, ok := stmt.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	if ident, ok := sel.X.(*ast.Ident); ok {
		if ident.Name != "t" && ident.Name != "b" {
			return false
		}
	}

	method := sel.Sel.Name
	return method == "Fatal" || method == "Fatalf" || method == "Error" || method == "Errorf"
}

// exprToString converts an expression to a string representation.
func (r *AssertUsageRule) exprToString(expr ast.Expr) string {
	result := api.ExprToString(expr)
	if result != "" {
		return result
	}

	switch e := expr.(type) {
	case *ast.BasicLit:
		return e.Value
	case *ast.CallExpr:
		if ident, ok := e.Fun.(*ast.Ident); ok && ident.Name == "len" && len(e.Args) == 1 {
			return "len(" + r.exprToString(e.Args[0]) + ")"
		}
	}

	return ""
}

// checkAssertEqualLen detects "assert.Equal(t, n, len(x))" patterns that should use assert.Len.
func (r *AssertUsageRule) checkAssertEqualLen(fset *token.FileSet, call *ast.CallExpr) *violation.Violation {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok || pkgIdent.Name != "assert" || sel.Sel.Name != "Equal" {
		return nil
	}

	if len(call.Args) < 3 {
		return nil
	}

	arg1 := call.Args[1]
	arg2 := call.Args[2]

	var lenCall *ast.CallExpr
	var expected ast.Expr
	var object ast.Expr

	if lc, ok := arg1.(*ast.CallExpr); ok && isLenCall(lc) {
		lenCall = lc
		expected = arg2
		object = lenCall.Args[0]
	} else if lc, ok := arg2.(*ast.CallExpr); ok && isLenCall(lc) {
		lenCall = lc
		expected = arg1
		object = lenCall.Args[0]
	}

	if lenCall == nil {
		return nil
	}

	pos := fset.Position(call.Pos())
	objectStr := r.exprToString(object)
	expectedStr := r.exprToString(expected)

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use assert.Len(t, %s, %s) instead of assert.Equal with len()", objectStr, expectedStr),
		Position: pos,
		Severity: r.Severity(),
	}
}

// extractNilComparison extracts the non-nil expression from a nil comparison.
// Returns nil if the condition is not a nil comparison with the given operator.
func extractNilComparison(cond ast.Expr, op token.Token) ast.Expr {
	binExpr, ok := cond.(*ast.BinaryExpr)
	if !ok || binExpr.Op != op {
		return nil
	}

	if isNilIdent(binExpr.Y) {
		return binExpr.X
	}
	if isNilIdent(binExpr.X) {
		return binExpr.Y
	}
	return nil
}

// isNilIdent checks if an expression is the nil identifier.
func isNilIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == nilLiteral
}

// isLenCall checks if a call expression is a len() call.
func isLenCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	return ok && ident.Name == "len" && len(call.Args) == 1
}
