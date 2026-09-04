// Package codequality contains rules for code quality and organization.
package codequality

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"unicode"

	"github.com/retroenv/retrogolib/set"
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
	return "Declarations should be ordered: types before typed constants; exported types, exported constructors, exported methods, unexported methods on exported types, exported functions, unexported types, unexported constructors, methods on unexported types, unexported functions. Exception: unexported type dependencies may appear directly before an exported type when that type uses them."
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

		if r.isAllowedDependencyTypePair(previous, current, decls) {
			continue
		}

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

	violations = append(violations, r.checkTypedConstantOrdering(fset, file)...)
	violations = append(violations, r.checkDependencyOrdering(fset, decls)...)
	return violations
}

func (r *FuncOrderRule) checkTypedConstantOrdering(fset *token.FileSet, file *ast.File) []violation.Violation {
	typePositions := namedTypePositions(file)
	var violations []violation.Violation

	for _, declaration := range file.Decls {
		constantDeclaration, ok := declaration.(*ast.GenDecl)
		if !ok || constantDeclaration.Tok != token.CONST {
			continue
		}

		for _, specification := range constantDeclaration.Specs {
			constant, ok := specification.(*ast.ValueSpec)
			if !ok || constant.Type == nil || len(constant.Names) == 0 {
				continue
			}

			for _, typeName := range referencedTypeNames(constant.Type) {
				typePosition, ok := typePositions[typeName]
				if !ok || typePosition < constant.Pos() {
					continue
				}

				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "type " + typeName + " should be declared before constant " + constant.Names[0].Name,
					Position: fset.Position(typePosition),
					Severity: r.Severity(),
				})
			}
		}
	}

	return violations
}

func namedTypePositions(file *ast.File) map[string]token.Pos {
	positions := make(map[string]token.Pos)

	for _, declaration := range file.Decls {
		typeDeclaration, ok := declaration.(*ast.GenDecl)
		if !ok || typeDeclaration.Tok != token.TYPE {
			continue
		}

		for _, specification := range typeDeclaration.Specs {
			typeSpec, ok := specification.(*ast.TypeSpec)
			if ok {
				positions[typeSpec.Name.Name] = typeSpec.Pos()
			}
		}
	}

	return positions
}

func (r *FuncOrderRule) checkDependencyOrdering(fset *token.FileSet, decls []declInfo) []violation.Violation {
	typeDeclByName := make(map[string]declInfo, len(decls))
	for _, decl := range decls {
		if decl.typ == nil {
			continue
		}
		typeDeclByName[decl.name] = decl
	}

	var violations []violation.Violation
	for _, decl := range decls {
		if decl.category != categoryExportedType || decl.typ == nil {
			continue
		}
		for _, depName := range referencedTypeNames(decl.typ.Type) {
			depDecl, ok := typeDeclByName[depName]
			if !ok || !isUnexportedType(depDecl.typ) {
				continue
			}
			if depDecl.pos > decl.pos && r.exportedTypeUsesDependency(decl, depName, decls) {
				pos := fset.Position(depDecl.pos)
				violations = append(violations, violation.Violation{
					Rule:     r.Name(),
					Message:  "type dependency " + depName + " should be declared before exported type " + decl.name,
					Position: pos,
					Severity: r.Severity(),
				})
			}
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
		fn:   funcDecl,
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
		typ:  typeSpec,
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

func (r *FuncOrderRule) isAllowedDependencyTypePair(previous, current declInfo, decls []declInfo) bool {
	if previous.category != categoryUnexportedType || current.category != categoryExportedType {
		return false
	}
	if previous.typ == nil || current.typ == nil {
		return false
	}
	if !isUnexportedType(previous.typ) {
		return false
	}
	return r.exportedTypeUsesDependency(current, previous.name, decls)
}

func (r *FuncOrderRule) exportedTypeUsesDependency(exportedType declInfo, depName string, decls []declInfo) bool {
	if exportedType.typ == nil || exportedType.category != categoryExportedType {
		return false
	}
	if typeReferencesIdent(exportedType.typ.Type, depName) {
		return true
	}
	for _, decl := range decls {
		if decl.fn == nil || decl.fn.Recv == nil {
			continue
		}
		if receiverTypeName(decl.fn.Recv) != exportedType.name {
			continue
		}
		if functionSignatureReferencesIdent(decl.fn.Type, depName) {
			return true
		}
	}
	return false
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
	typ      *ast.TypeSpec
	fn       *ast.FuncDecl
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

func isUnexportedType(typeSpec *ast.TypeSpec) bool {
	if typeSpec == nil {
		return false
	}
	return !isExported(typeSpec.Name.Name)
}

func typeReferencesIdent(expr ast.Expr, ident string) bool {
	found := false
	ast.Inspect(expr, func(node ast.Node) bool {
		id, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		if id.Name == ident {
			found = true
			return false
		}
		return true
	})
	return found
}

func referencedTypeNames(expr ast.Expr) []string {
	seen := set.New[string]()
	ast.Inspect(expr, func(node ast.Node) bool {
		id, ok := node.(*ast.Ident)
		if !ok {
			return true
		}
		if id.Name == "" {
			return true
		}
		seen.Add(id.Name)
		return true
	})

	return set.Sorted(seen)
}

func functionSignatureReferencesIdent(fnType *ast.FuncType, ident string) bool {
	if fnType == nil {
		return false
	}
	if fieldListReferencesIdent(fnType.Params, ident) {
		return true
	}
	return fieldListReferencesIdent(fnType.Results, ident)
}

func fieldListReferencesIdent(fields *ast.FieldList, ident string) bool {
	if fields == nil {
		return false
	}
	for _, field := range fields.List {
		if typeReferencesIdent(field.Type, ident) {
			return true
		}
	}
	return false
}

func receiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	return baseTypeName(recv.List[0].Type)
}

func baseTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return baseTypeName(t.X)
	case *ast.IndexExpr:
		return baseTypeName(t.X)
	case *ast.IndexListExpr:
		return baseTypeName(t.X)
	default:
		return ""
	}
}
