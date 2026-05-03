// Package codequality contains rules for code quality and organization.
package codequality

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"unicode"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// FuncOrderRule detects functions and methods that are not in the proper order.
type FuncOrderRule struct{}

// NewFuncOrderRule creates a new FuncOrderRule.
func NewFuncOrderRule() *FuncOrderRule {
	return &FuncOrderRule{}
}

// Name returns the rule name.
func (r *FuncOrderRule) Name() string {
	return "codequality-funcorder"
}

// Description returns the rule description.
func (r *FuncOrderRule) Description() string {
	return "Declarations should be ordered: exported types, exported constructors, exported methods, unexported methods on exported types, exported functions, unexported types, unexported constructors, methods on unexported types, unexported functions"
}

// Severity returns the default severity.
func (r *FuncOrderRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *FuncOrderRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes a file for violations.
func (r *FuncOrderRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	if isTestFile(fset, file) {
		return nil
	}

	var violations []violation.Violation
	var decls []declInfo

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if info, ok := r.categorizeFunc(d); ok {
				decls = append(decls, info)
			}
		case *ast.GenDecl:
			if info, ok := r.categorizeTypeDecl(d); ok {
				decls = append(decls, info)
			}
		}
	}

	for i := 1; i < len(decls); i++ {
		current := decls[i]
		previous := decls[i-1]

		if current.category < previous.category {
			pos := fset.Position(current.pos)
			violations = append(violations, violation.Violation{
				Rule:     r.Name(),
				Message:  r.formatMessage(current.name, current.category, previous.category),
				Position: pos,
				Severity: r.Severity(),
			})
		}
	}

	return violations
}

// categorizeFunc determines the category of a function/method.
func (r *FuncOrderRule) categorizeFunc(funcDecl *ast.FuncDecl) (declInfo, bool) {
	name := funcDecl.Name.Name
	isMethod := funcDecl.Recv != nil

	info := declInfo{
		name: name,
		pos:  funcDecl.Pos(),
	}

	if isInitFunction(name, isMethod) {
		return declInfo{}, false
	}

	if isMethod {
		info.category = categorizeMethod(name, funcDecl.Recv)
		return info, true
	}

	if isConstructor(funcDecl) {
		if isExported(name) {
			info.category = categoryExportedConstructor
		} else {
			info.category = categoryUnexportedConstructor
		}
		return info, true
	}

	if isExported(name) {
		info.category = categoryExportedFunction
	} else {
		info.category = categoryUnexportedFunction
	}
	return info, true
}

// categorizeTypeDecl determines the category of a type declaration.
// It returns false if the GenDecl is not a type declaration.
func (r *FuncOrderRule) categorizeTypeDecl(genDecl *ast.GenDecl) (declInfo, bool) {
	if genDecl.Tok != token.TYPE {
		return declInfo{}, false
	}
	if len(genDecl.Specs) == 0 {
		return declInfo{}, false
	}

	typeSpec, ok := genDecl.Specs[0].(*ast.TypeSpec)
	if !ok {
		return declInfo{}, false
	}

	name := typeSpec.Name.Name
	info := declInfo{
		name: name,
		pos:  genDecl.Pos(),
	}

	if isExported(name) {
		info.category = categoryExportedType
	} else {
		info.category = categoryUnexportedType
	}
	return info, true
}

// formatMessage creates a descriptive message for the violation.
func (r *FuncOrderRule) formatMessage(name string, current, previous funcCategory) string {
	return name + " (" + current.String() + ") should come before " + previous.String()
}

// funcCategory represents the category of a function/method or type declaration.
type funcCategory int

const (
	categoryExportedType funcCategory = iota
	categoryExportedConstructor
	categoryExportedMethod
	categoryUnexportedMethod
	categoryExportedFunction
	categoryUnexportedType
	categoryUnexportedConstructor
	categoryUnexportedTypeMethod
	categoryUnexportedFunction
)

// declInfo holds information about a top-level declaration.
type declInfo struct {
	name     string
	category funcCategory
	pos      token.Pos
}

func (c funcCategory) String() string {
	switch c {
	case categoryExportedType:
		return "exported type"
	case categoryExportedConstructor:
		return "exported constructor"
	case categoryExportedMethod:
		return "exported method"
	case categoryUnexportedMethod:
		return "unexported method"
	case categoryExportedFunction:
		return "exported function"
	case categoryUnexportedType:
		return "unexported type"
	case categoryUnexportedConstructor:
		return "unexported constructor"
	case categoryUnexportedTypeMethod:
		return "method on unexported type"
	case categoryUnexportedFunction:
		return "unexported function"
	default:
		return "unknown"
	}
}

func categorizeMethod(name string, recv *ast.FieldList) funcCategory {
	receiverExported := true
	if recv != nil && len(recv.List) > 0 {
		receiverExported = isExportedReceiverType(recv.List[0].Type)
	}

	if !receiverExported {
		return categoryUnexportedTypeMethod
	}
	if isExported(name) {
		return categoryExportedMethod
	}
	return categoryUnexportedMethod
}

func isTestFile(fset *token.FileSet, file *ast.File) bool {
	filename := fset.Position(file.Package).Filename
	return filepath.Ext(filename) == ".go" && len(filename) >= len("_test.go") &&
		filename[len(filename)-len("_test.go"):] == "_test.go"
}

func isInitFunction(name string, isMethod bool) bool {
	return name == "init" && !isMethod
}

func isExported(name string) bool {
	return len(name) > 0 && unicode.IsUpper(rune(name[0]))
}

func isConstructor(funcDecl *ast.FuncDecl) bool {
	return isConstructorName(funcDecl.Name.Name) && returnsNamedType(funcDecl.Type.Results)
}

func isConstructorName(name string) bool {
	switch {
	case name == "New":
		return true
	case len(name) > len("New") && name[:len("New")] == "New":
		return isConstructorBoundary(rune(name[len("New")]))
	case name == "new":
		return true
	case len(name) > len("new") && name[:len("new")] == "new":
		return isConstructorBoundary(rune(name[len("new")]))
	default:
		return false
	}
}

func isConstructorBoundary(r rune) bool {
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}

func returnsNamedType(results *ast.FieldList) bool {
	if results == nil {
		return false
	}

	for _, result := range results.List {
		if isNamedReturnType(result.Type) {
			return true
		}
	}
	return false
}

func isNamedReturnType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return !isPredeclaredType(t.Name)
	case *ast.StarExpr:
		return isNamedReturnType(t.X)
	case *ast.SelectorExpr:
		return true
	case *ast.IndexExpr:
		return isNamedReturnType(t.X)
	case *ast.IndexListExpr:
		return isNamedReturnType(t.X)
	default:
		return false
	}
}

func isPredeclaredType(name string) bool {
	switch name {
	case "any", "bool", "byte", "comparable", "complex64", "complex128", "error", "float32", "float64",
		"int", "int8", "int16", "int32", "int64", "rune", "string",
		"uint", "uint8", "uint16", "uint32", "uint64", "uintptr":
		return true
	default:
		return false
	}
}

// isExportedReceiverType returns true if the receiver type expression refers to an exported type.
func isExportedReceiverType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return isExported(t.Name)
	case *ast.StarExpr:
		return isExportedReceiverType(t.X)
	case *ast.IndexExpr:
		return isExportedReceiverType(t.X)
	case *ast.IndexListExpr:
		return isExportedReceiverType(t.X)
	default:
		return true
	}
}
