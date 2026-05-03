package codequality

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// TypeStutterRule detects exported type names that stutter the package name.
// For example, in package "register", a type named "RegisterMeta" stutters
// because callers would write "register.RegisterMeta".
type TypeStutterRule struct{}

// NewTypeStutterRule creates a new TypeStutterRule.
func NewTypeStutterRule() *TypeStutterRule {
	return &TypeStutterRule{}
}

// Name returns the rule name.
func (r *TypeStutterRule) Name() string {
	return "codequality-type-stutter"
}

// Description returns the rule description.
func (r *TypeStutterRule) Description() string {
	return "Exported type name should not stutter the package name (e.g. register.RegisterMeta should be register.Meta)"
}

// Severity returns the default severity.
func (r *TypeStutterRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *TypeStutterRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for type name stutter violations.
func (r *TypeStutterRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	pkgName := file.Name.Name
	// Strip _test suffix from external test packages.
	pkgName = strings.TrimSuffix(pkgName, "_test")
	if len(pkgName) < 2 {
		return violations
	}

	pkgLower := strings.ToLower(pkgName)

	ast.Inspect(file, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		typeName := typeSpec.Name.Name
		if !ast.IsExported(typeName) {
			return true
		}

		if !strings.HasPrefix(strings.ToLower(typeName), pkgLower) {
			return true
		}

		// There must be a non-empty suffix after the package-name prefix.
		// A type named exactly after its package (e.g. Pipeline in package pipeline)
		// is a common, accepted Go pattern and should not be flagged.
		rest := typeName[len(pkgName):]
		if rest == "" {
			return true
		}

		// The character immediately after the package-name prefix must be
		// uppercase or non-letter (a CamelCase word boundary), otherwise
		// the match is coincidental (e.g. "Registered" in package "register").
		if unicode.IsLower(rune(rest[0])) {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "type \"" + typeName + "\" stutters the package name \"" + pkgName + "\"; consider \"" + rest + "\"",
			Position: fset.Position(typeSpec.Pos()),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}
