package codequality

import (
	"go/ast"
	"go/token"
	"sort"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const (
	conditionHeader = "if condition"
	signatureHeader = "function signature"
)

// BlockPaddingRule detects blocks that start or end with an empty line, and blocks that follow a multiline if
// condition or function signature without an empty line or a comment before the first statement.
type BlockPaddingRule struct{}

// NewBlockPaddingRule creates a new BlockPaddingRule.
func NewBlockPaddingRule() *BlockPaddingRule {
	return &BlockPaddingRule{}
}

// Name returns the rule name.
func (r *BlockPaddingRule) Name() string {
	return "codequality-block-padding"
}

// Description returns the rule description.
func (r *BlockPaddingRule) Description() string {
	return "Blocks should not start or end with an empty line, and bodies of multiline if conditions and " +
		"function signatures should start with an empty line or a comment"
}

// Severity returns the default severity.
func (r *BlockPaddingRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *BlockPaddingRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for blocks with empty lines at the start or end, and for bodies of multiline if
// conditions and function signatures that start directly with a statement.
func (r *BlockPaddingRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	headers := map[*ast.BlockStmt]string{}

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.IfStmt:
			markMultilineHeader(fset, headers, n.Body, n.Cond, conditionHeader)
		case *ast.FuncLit:
			markMultilineHeader(fset, headers, n.Body, n.Type, signatureHeader)
		case *ast.FuncDecl:
			markMultilineHeader(fset, headers, n.Body, n.Type, signatureHeader)
		case *ast.BlockStmt:
			violations = append(violations, r.checkBlock(fset, file, n, headers[n])...)
		}

		return true
	})

	return violations
}

// checkBlock checks the start and the end of a block. The header is the name of the multiline construct
// that opens the block, or empty if the construct is on one line.
func (r *BlockPaddingRule) checkBlock(fset *token.FileSet, file *ast.File, block *ast.BlockStmt,
	header string) []violation.Violation {

	if len(block.List) == 0 {
		return nil
	}

	var violations []violation.Violation

	openLine := fset.Position(block.Lbrace).Line
	closeLine := fset.Position(block.Rbrace).Line

	if header != "" {
		if fset.Position(block.List[0].Pos()).Line <= openLine+1 {
			violations = append(violations, r.violation(fset, block.Lbrace,
				"multiline "+header+" should be followed by an empty line or a comment"))
		}
	} else if line, ok := leadingLine(fset, file, block); ok && line > openLine+1 {
		violations = append(violations, r.violation(fset, block.Lbrace, "block should not start with an empty line"))
	}

	if closeLine-1 > trailingLine(fset, file, block) {
		violations = append(violations, r.violation(fset, block.Rbrace, "block should not end with an empty line"))
	}

	return violations
}

func (r *BlockPaddingRule) violation(fset *token.FileSet, pos token.Pos, message string) violation.Violation {
	return violation.Violation{
		Rule:     r.Name(),
		Message:  message,
		Position: fset.Position(pos),
		Severity: r.Severity(),
	}
}

// markMultilineHeader records the body of a construct whose header spans more than one line.
func markMultilineHeader(fset *token.FileSet, headers map[*ast.BlockStmt]string, body *ast.BlockStmt, header ast.Node,
	name string) {

	if body == nil {
		return
	}

	if fset.Position(header.Pos()).Line != fset.Position(header.End()).Line {
		headers[body] = name
	}
}

// leadingLine returns the line of the first statement or comment of a non-empty block.
// A comment that starts on the line of the opening brace and ends on a later line is accepted as is, and the
// result is reported as not available.
func leadingLine(fset *token.FileSet, file *ast.File, block *ast.BlockStmt) (int, bool) {
	openLine := fset.Position(block.Lbrace).Line
	first := block.List[0]

	for _, group := range commentsAfter(file, block.Lbrace) {
		if group.Pos() >= first.Pos() {
			break
		}

		startLine := fset.Position(group.Pos()).Line
		if startLine != openLine {
			return startLine, true
		}

		if fset.Position(group.End()).Line != openLine {
			return 0, false
		}
	}

	return fset.Position(first.Pos()).Line, true
}

// trailingLine returns the last line of the last statement or comment of a non-empty block.
// A comment that ends on the line of the closing brace is not part of the block content.
func trailingLine(fset *token.FileSet, file *ast.File, block *ast.BlockStmt) int {
	last := block.List[len(block.List)-1]
	line := fset.Position(last.End()).Line
	closeLine := fset.Position(block.Rbrace).Line

	for _, group := range commentsAfter(file, last.End()) {
		if group.End() > block.Rbrace {
			break
		}

		if endLine := fset.Position(group.End()).Line; endLine != closeLine {
			line = max(line, endLine)
		}
	}

	return line
}

// commentsAfter returns the comment groups that start after a position.
func commentsAfter(file *ast.File, pos token.Pos) []*ast.CommentGroup {
	first := sort.Search(len(file.Comments), func(i int) bool {
		return file.Comments[i].Pos() > pos
	})

	return file.Comments[first:]
}
