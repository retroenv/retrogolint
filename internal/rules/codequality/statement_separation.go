package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// minStatementsBeforeBranch is the number of statements that must be before a return or branch statement
// before the statement needs an empty line. A block of three statements can stay compact.
const minStatementsBeforeBranch = 3

// StatementSeparationRule detects multiline control statements and return or branch statements that are not
// separated from adjacent statements by an empty line.
type StatementSeparationRule struct{}

// NewStatementSeparationRule creates a new StatementSeparationRule.
func NewStatementSeparationRule() *StatementSeparationRule {
	return &StatementSeparationRule{}
}

// Name returns the rule name.
func (r *StatementSeparationRule) Name() string {
	return "codequality-statement-separation"
}

// Description returns the rule description.
func (r *StatementSeparationRule) Description() string {
	return "Multiline control statements and return or branch statements should be separated from adjacent " +
		"statements by an empty line"
}

// Severity returns the default severity.
func (r *StatementSeparationRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *StatementSeparationRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for statements that directly follow a multiline control statement, and for return and
// branch statements that directly follow other statements in a block of three or more statements.
func (r *StatementSeparationRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(node ast.Node) bool {
		list := statementList(node)

		for index := 1; index < len(list); index++ {
			message := separationMessage(fset, list, index)
			if message == "" || hasEmptyLine(fset, file, list[index-1].End(), list[index].Pos()) {
				continue
			}

			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  message,
				Position: fset.Position(list[index].Pos()),
				Severity: r.Severity(),
			})
		}

		return true
	})

	return violations
}

// statementList returns the statements of a block, switch case, or select case.
func statementList(node ast.Node) []ast.Stmt {
	switch n := node.(type) {
	case *ast.BlockStmt:
		return n.List
	case *ast.CaseClause:
		return n.Body
	case *ast.CommClause:
		return n.Body
	default:
		return nil
	}
}

// separationMessage returns the violation message when the statement at index needs an empty line before it.
// It returns an empty string when the statement does not need an empty line.
func separationMessage(fset *token.FileSet, list []ast.Stmt, index int) string {
	previous := list[index-1]
	next := list[index]

	if name, ok := controlStatementName(previous); ok && isMultilineStatement(fset, previous) {
		if isGuardClause(previous) && isIfStatement(next) {
			return ""
		}

		return "multiline " + name + " statement should be followed by an empty line"
	}

	if index < minStatementsBeforeBranch {
		return ""
	}

	if name, ok := branchName(next); ok {
		return name + " statement should be separated from the previous statements by an empty line"
	}

	return ""
}

// controlStatementName returns the keyword of an if, for, range, switch, or select statement.
func controlStatementName(stmt ast.Stmt) (string, bool) {
	switch stmt.(type) {
	case *ast.IfStmt:
		return "if", true
	case *ast.ForStmt:
		return "for", true
	case *ast.RangeStmt:
		return "range", true
	case *ast.SwitchStmt, *ast.TypeSwitchStmt:
		return "switch", true
	case *ast.SelectStmt:
		return "select", true
	default:
		return "", false
	}
}

// branchName returns the keyword of a return, break, continue, or goto statement.
// A fallthrough statement belongs to its case body and is not a branch statement here.
func branchName(stmt ast.Stmt) (string, bool) {
	switch s := stmt.(type) {
	case *ast.ReturnStmt:
		return "return", true

	case *ast.BranchStmt:
		if s.Tok != token.FALLTHROUGH {
			return s.Tok.String(), true
		}
	}

	return "", false
}

// isGuardClause reports whether a statement is an if statement without else branch whose body ends with a
// return, break, continue, goto, or panic call. A guard clause can stay attached to a following if statement.
func isGuardClause(stmt ast.Stmt) bool {
	ifStmt, ok := stmt.(*ast.IfStmt)
	if !ok || ifStmt.Else != nil || len(ifStmt.Body.List) == 0 {
		return false
	}

	last := ifStmt.Body.List[len(ifStmt.Body.List)-1]
	if _, ok := branchName(last); ok {
		return true
	}

	exprStmt, ok := last.(*ast.ExprStmt)
	if !ok {
		return false
	}

	call, ok := exprStmt.X.(*ast.CallExpr)
	if !ok {
		return false
	}

	ident, ok := call.Fun.(*ast.Ident)

	return ok && ident.Name == "panic"
}

// isIfStatement reports whether a statement is an if statement.
func isIfStatement(stmt ast.Stmt) bool {
	_, ok := stmt.(*ast.IfStmt)
	return ok
}

// isMultilineStatement reports whether a statement spans more than one line.
func isMultilineStatement(fset *token.FileSet, stmt ast.Stmt) bool {
	return fset.Position(stmt.Pos()).Line != fset.Position(stmt.End()).Line
}
