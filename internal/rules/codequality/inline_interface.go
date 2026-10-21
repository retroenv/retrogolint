package codequality

import (
	"go/ast"
	"go/token"

	"github.com/retroenv/retrogolib/set"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// InlineInterfaceRule detects non-empty interface types that are written inline.
// For example, "x.(interface{ Resolve() })" should use a named interface type.
// Empty interfaces, named interface declarations, and type parameter constraints are not reported.
type InlineInterfaceRule struct{}

// NewInlineInterfaceRule creates a new InlineInterfaceRule.
func NewInlineInterfaceRule() *InlineInterfaceRule {
	return &InlineInterfaceRule{}
}

// Name returns the rule name.
func (r *InlineInterfaceRule) Name() string {
	return "codequality-inline-interface"
}

// Description returns the rule description.
func (r *InlineInterfaceRule) Description() string {
	return "Non-empty interface types should be declared as named types instead of written inline"
}

// Severity returns the default severity.
func (r *InlineInterfaceRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *InlineInterfaceRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for non-empty interface types that are written inline.
func (r *InlineInterfaceRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	// ast.Inspect visits a declaration before its children, so the exempt set
	// is complete when the interface nodes are reached.
	exempt := set.New[*ast.InterfaceType]()

	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.TypeSpec:
			addExemptInterface(exempt, n.Type)
			addExemptConstraints(exempt, n.TypeParams)

		case *ast.FuncDecl:
			addExemptConstraints(exempt, n.Type.TypeParams)

		case *ast.InterfaceType:
			if exempt.Contains(n) || len(n.Methods.List) == 0 {
				return true
			}

			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  "inline interface type should be declared as a named type",
				Position: fset.Position(n.Pos()),
				Severity: r.Severity(),
			})
		}

		return true
	})

	return violations
}

// addExemptInterface marks an interface type that is the direct type of a named declaration.
func addExemptInterface(exempt set.Set[*ast.InterfaceType], expr ast.Expr) {
	if iface, ok := expr.(*ast.InterfaceType); ok {
		exempt.Add(iface)
	}
}

// addExemptConstraints marks the interface types used as type parameter constraints.
func addExemptConstraints(exempt set.Set[*ast.InterfaceType], params *ast.FieldList) {
	if params == nil {
		return
	}

	for _, field := range params.List {
		addExemptInterface(exempt, field.Type)
	}
}
