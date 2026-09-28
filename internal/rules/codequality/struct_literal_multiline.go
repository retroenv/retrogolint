package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// StructLiteralMultilineRule detects multi-field struct literals with fields on the same line as the opening brace or another field.
type StructLiteralMultilineRule struct{}

// NewStructLiteralMultilineRule creates a new StructLiteralMultilineRule.
func NewStructLiteralMultilineRule() *StructLiteralMultilineRule {
	return &StructLiteralMultilineRule{}
}

// Name returns the rule name.
func (r *StructLiteralMultilineRule) Name() string {
	return "codequality-struct-literal-multiline"
}

// Description returns the rule description.
func (r *StructLiteralMultilineRule) Description() string {
	return "Struct literals with multiple fields should start fields on a new line and use one line per field"
}

// Severity returns the default severity.
func (r *StructLiteralMultilineRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *StructLiteralMultilineRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for multi-field struct literals that do not put the first field on a new line or put two fields on one line.
func (r *StructLiteralMultilineRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok || !isKeyedStructLiteral(literal) {
			return true
		}

		fieldPos := sharedLineFieldPos(fset, literal)
		if !fieldPos.IsValid() {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "struct literal fields should start on a new line and use one line per field",
			Position: fset.Position(fieldPos),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}

// sharedLineFieldPos returns the position of a field that shares a line with the opening brace or the previous field.
func sharedLineFieldPos(fset *token.FileSet, literal *ast.CompositeLit) token.Pos {
	if fset.Position(literal.Lbrace).Line == fset.Position(literal.Elts[0].Pos()).Line {
		return literal.Elts[0].Pos()
	}

	for i := 1; i < len(literal.Elts); i++ {
		previousLine := fset.Position(literal.Elts[i-1].End()).Line
		currentLine := fset.Position(literal.Elts[i].Pos()).Line
		if previousLine == currentLine {
			return literal.Elts[i].Pos()
		}
	}

	return token.NoPos
}

func isKeyedStructLiteral(literal *ast.CompositeLit) bool {
	if len(literal.Elts) < 2 {
		return false
	}

	switch literal.Type.(type) {
	case nil, *ast.ArrayType, *ast.MapType:
		return false
	}

	for _, element := range literal.Elts {
		keyValue, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return false
		}
		if _, ok := keyValue.Key.(*ast.Ident); !ok {
			return false
		}
	}

	return true
}
