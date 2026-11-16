package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// InterfaceMultilineRule detects non-empty interface type declarations on one line.
// For example, "type Device interface{ TransferState(*Archive) error }" should use multiple lines.
// Empty interfaces and interface constraints written inline are not reported.
type InterfaceMultilineRule struct{}

// NewInterfaceMultilineRule creates a new InterfaceMultilineRule.
func NewInterfaceMultilineRule() *InterfaceMultilineRule {
	return &InterfaceMultilineRule{}
}

// Name returns the rule name.
func (r *InterfaceMultilineRule) Name() string {
	return "codequality-interface-multiline"
}

// Description returns the rule description.
func (r *InterfaceMultilineRule) Description() string {
	return "Non-empty interface type declarations should use multiple lines"
}

// Severity returns the default severity.
func (r *InterfaceMultilineRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *InterfaceMultilineRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for interface type declarations that are written on one line.
func (r *InterfaceMultilineRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	ast.Inspect(file, func(node ast.Node) bool {
		typeSpec, ok := node.(*ast.TypeSpec)
		if !ok {
			return true
		}

		iface, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok || len(iface.Methods.List) == 0 {
			return true
		}

		if fset.Position(iface.Interface).Line != fset.Position(iface.Methods.Closing).Line {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "interface type declaration should use multiple lines",
			Position: fset.Position(typeSpec.Name.Pos()),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}
