package codequality

import (
	"go/ast"
	"go/token"
	"sort"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// SwitchCaseSeparationRule detects multiline switch and select case bodies that are not separated from
// adjacent cases by an empty line.
type SwitchCaseSeparationRule struct{}

// NewSwitchCaseSeparationRule creates a new SwitchCaseSeparationRule.
func NewSwitchCaseSeparationRule() *SwitchCaseSeparationRule {
	return &SwitchCaseSeparationRule{}
}

// Name returns the rule name.
func (r *SwitchCaseSeparationRule) Name() string {
	return "codequality-switch-case-separation"
}

// Description returns the rule description.
func (r *SwitchCaseSeparationRule) Description() string {
	return "Switch and select cases with multiline bodies should be separated from adjacent cases by an empty line"
}

// Severity returns the default severity.
func (r *SwitchCaseSeparationRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *SwitchCaseSeparationRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for case labels that directly follow or precede a multiline case body without
// an empty line.
func (r *SwitchCaseSeparationRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(node ast.Node) bool {
		clauses, ok := caseClauses(node)
		if !ok {
			return true
		}

		for index := 0; index+1 < len(clauses); index++ {
			if !needsSeparator(fset, clauses, index) || hasEmptyLine(fset, file, clauses[index], clauses[index+1]) {
				continue
			}

			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  "multiline case bodies should be separated from adjacent cases by an empty line",
				Position: fset.Position(clauses[index+1].label),
				Severity: r.Severity(),
			})
		}

		return true
	})

	return violations
}

// caseClause is a case of a switch, type switch, or select statement.
type caseClause struct {
	label token.Pos
	body  []ast.Stmt
}

// caseClauses returns the cases of a switch, type switch, or select statement.
func caseClauses(node ast.Node) ([]caseClause, bool) {
	var list []ast.Stmt

	switch s := node.(type) {
	case *ast.SwitchStmt:
		list = s.Body.List
	case *ast.TypeSwitchStmt:
		list = s.Body.List
	case *ast.SelectStmt:
		list = s.Body.List
	default:
		return nil, false
	}

	clauses := make([]caseClause, 0, len(list))
	for _, stmt := range list {
		switch clause := stmt.(type) {
		case *ast.CaseClause:
			clauses = append(clauses, caseClause{
				label: clause.Case,
				body:  clause.Body,
			})

		case *ast.CommClause:
			clauses = append(clauses, caseClause{
				label: clause.Case,
				body:  clause.Body,
			})
		}
	}

	return clauses, true
}

// needsSeparator reports whether the case at index must be separated from the next case by an empty line.
// An empty case body groups its label with the next case, and a fallthrough body keeps the next case attached.
func needsSeparator(fset *token.FileSet, clauses []caseClause, index int) bool {
	previous := clauses[index].body
	if len(previous) == 0 || isFallthrough(previous[len(previous)-1]) {
		return false
	}

	return isMultiline(fset, previous) || isMultiline(fset, groupBody(clauses, index+1))
}

// groupBody returns the body of the first case at or after start that has statements.
// Cases with empty bodies before that case share its body.
func groupBody(clauses []caseClause, start int) []ast.Stmt {
	for _, clause := range clauses[start:] {
		if len(clause.body) > 0 {
			return clause.body
		}
	}

	return nil
}

// isMultiline reports whether a case body spans more than one line.
func isMultiline(fset *token.FileSet, body []ast.Stmt) bool {
	if len(body) == 0 {
		return false
	}

	firstLine := fset.Position(body[0].Pos()).Line
	lastLine := fset.Position(body[len(body)-1].End()).Line
	return firstLine != lastLine
}

// isFallthrough reports whether a statement is a fallthrough statement.
func isFallthrough(stmt ast.Stmt) bool {
	branch, ok := stmt.(*ast.BranchStmt)
	return ok && branch.Tok == token.FALLTHROUGH
}

// hasEmptyLine reports whether an empty line is between the body of a case and the label of the next case.
// Comment lines in that range are not empty lines.
func hasEmptyLine(fset *token.FileSet, file *ast.File, previous, next caseClause) bool {
	bodyEnd := previous.body[len(previous.body)-1].End()
	firstGapLine := fset.Position(bodyEnd).Line + 1
	labelLine := fset.Position(next.label).Line
	gapLines := labelLine - firstGapLine

	commentLines := 0
	first := sort.Search(len(file.Comments), func(i int) bool {
		return file.Comments[i].Pos() > bodyEnd
	})
	for _, group := range file.Comments[first:] {
		if group.End() > next.label {
			break
		}

		startLine := max(fset.Position(group.Pos()).Line, firstGapLine)
		endLine := min(fset.Position(group.End()).Line, labelLine-1)
		if endLine >= startLine {
			commentLines += endLine - startLine + 1
		}
	}

	return commentLines < gapLines
}
