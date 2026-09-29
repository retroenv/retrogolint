// Package codequality contains rules for code quality and organization.
package codequality

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const maxFunctionDeclarationColumns = 120

// FuncSignatureLayoutRule detects function signatures that do not use the available line width.
type FuncSignatureLayoutRule struct{}

// NewFuncSignatureLayoutRule creates a new FuncSignatureLayoutRule.
func NewFuncSignatureLayoutRule() *FuncSignatureLayoutRule {
	return &FuncSignatureLayoutRule{}
}

// Name returns the rule name.
func (r *FuncSignatureLayoutRule) Name() string {
	return "codequality-func-signature-layout"
}

// Description returns the rule description.
func (r *FuncSignatureLayoutRule) Description() string {
	return "Function signatures should use the available 120 columns and wrap after parameter commas"
}

// Severity returns the default severity.
func (r *FuncSignatureLayoutRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *FuncSignatureLayoutRule) Category() string {
	return api.CategoryCodeQuality
}

// Check analyzes function, method, and function type signature layout.
func (r *FuncSignatureLayoutRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	for _, declaration := range file.Decls {
		switch decl := declaration.(type) {
		case *ast.FuncDecl:
			violations = append(violations, r.checkFunctionDeclaration(fset, file, decl)...)
		case *ast.GenDecl:
			violations = append(violations, r.checkTypeDeclarations(fset, file, decl)...)
		}
	}

	return violations
}

// checkFunctionDeclaration returns a violation when a function or method signature has an invalid layout.
func (r *FuncSignatureLayoutRule) checkFunctionDeclaration(fset *token.FileSet, file *ast.File,
	function *ast.FuncDecl) []violation.Violation {

	return r.checkSignature(fset, file, signatureTarget{
		signature:  buildFunctionSignature(function),
		funcType:   function.Type,
		receiver:   function.Recv,
		typeParams: function.Type.TypeParams,
		end:        signatureEnd(function.Type, function.Body),
		position:   function.Name.Pos(),
		name:       function.Name.Name,
	})
}

// checkTypeDeclarations returns violations for function type declarations in a declaration group.
func (r *FuncSignatureLayoutRule) checkTypeDeclarations(fset *token.FileSet, file *ast.File,
	declaration *ast.GenDecl) []violation.Violation {

	var violations []violation.Violation

	for _, specification := range declaration.Specs {
		typeSpec, ok := specification.(*ast.TypeSpec)
		if !ok {
			continue
		}
		funcType, ok := typeSpec.Type.(*ast.FuncType)
		if !ok {
			continue
		}

		violations = append(violations, r.checkSignature(fset, file, signatureTarget{
			signature:  buildTypeSignature(fset, declaration, typeSpec, funcType),
			funcType:   funcType,
			typeParams: typeSpec.TypeParams,
			end:        signatureEnd(funcType, nil),
			position:   typeSpec.Name.Pos(),
			name:       typeSpec.Name.Name,
		})...)
	}

	return violations
}

// checkSignature returns a violation when the target signature has an invalid layout.
func (r *FuncSignatureLayoutRule) checkSignature(fset *token.FileSet, file *ast.File,
	target signatureTarget) []violation.Violation {

	if hasSignatureComments(file, target.funcType.Func, target.end) ||
		target.signature.hasValidLayout(fset, target.funcType, target.receiver, target.typeParams, target.end) {

		return nil
	}

	return []violation.Violation{{
		Rule:     r.Name(),
		Message:  target.name + ": function signature should use available 120 columns and wrap after parameter commas",
		Position: fset.Position(target.position),
		Severity: r.Severity(),
	}}
}

// signatureTarget identifies one function signature to check.
type signatureTarget struct {
	signature  functionSignature
	funcType   *ast.FuncType
	receiver   *ast.FieldList
	typeParams *ast.FieldList
	end        token.Pos

	name     string
	position token.Pos
}

type functionSignature struct {
	prefixWidth       int
	continuationWidth int
	parameters        []string
	suffix            string
}

type functionParameterPart struct {
	position token.Pos
	end      token.Pos
}

func (signature functionSignature) hasValidLayout(fset *token.FileSet, funcType *ast.FuncType, receiver,
	typeParams *ast.FieldList, end token.Pos) bool {

	if hasMultilineSignatureField(fset, receiver, typeParams, funcType) {
		return true
	}

	startLine := fset.Position(funcType.Func).Line
	endLine := fset.Position(end).Line
	fullWidth := signature.prefixWidth + len(strings.Join(signature.parameters, ", ")) + len(signature.suffix)

	if fullWidth <= maxFunctionDeclarationColumns {
		return startLine == endLine
	}

	if startLine == endLine {
		return signature.requiresIndivisibleOverrun()
	}
	if fset.Position(funcType.Params.Opening).Line != startLine {
		return false
	}

	parameters := functionParameterParts(funcType.Params.List)
	expectedLines := signature.parameterLineIndexes()
	if len(parameters) != len(expectedLines) {
		return false
	}

	for index, parameter := range parameters {
		if fset.Position(parameter.position).Line != fset.Position(parameter.end).Line {
			return false
		}
		if fset.Position(parameter.position).Line != startLine+expectedLines[index] {
			return false
		}
	}

	lastParameterLine := fset.Position(funcType.Params.Opening).Line
	if len(parameters) > 0 {
		lastParameterLine = fset.Position(parameters[len(parameters)-1].end).Line
	}
	if fset.Position(funcType.Params.Closing).Line != lastParameterLine {
		return false
	}

	return signature.hasValidResultLayout(fset, funcType, lastParameterLine)
}

func (signature functionSignature) requiresIndivisibleOverrun() bool {
	if len(signature.parameters) == 0 {
		return true
	}

	lastParameter := signature.parameters[len(signature.parameters)-1]
	return signature.continuationWidth+len(lastParameter)+len(signature.suffix) > maxFunctionDeclarationColumns
}

func (signature functionSignature) parameterLineIndexes() []int {
	lines := make([]int, len(signature.parameters))
	lineIndex := 0
	lineWidth := signature.prefixWidth
	parametersOnLine := 0

	for index, parameter := range signature.parameters {
		separatorWidth := 0
		if parametersOnLine > 0 {
			separatorWidth = 1
		}
		endingWidth := 1
		if index == len(signature.parameters)-1 {
			endingWidth = len(signature.parameterClosingSuffix(parameter))
		}

		candidateWidth := lineWidth + separatorWidth + len(parameter) + endingWidth
		if candidateWidth > maxFunctionDeclarationColumns && (lineIndex == 0 || parametersOnLine > 0) {
			lineIndex++
			lineWidth = signature.continuationWidth
			parametersOnLine = 0
			separatorWidth = 0
		}

		lines[index] = lineIndex
		lineWidth += separatorWidth + len(parameter) + endingWidth
		parametersOnLine++
	}

	return lines
}

func (signature functionSignature) parameterClosingSuffix(parameter string) string {
	if signature.suffix == ")" || signature.suffix == ") {" {
		return signature.suffix
	}
	if signature.continuationWidth+len(parameter)+len(signature.suffix) <= maxFunctionDeclarationColumns {
		return signature.suffix
	}

	return ")"
}

func (signature functionSignature) hasValidResultLayout(fset *token.FileSet, funcType *ast.FuncType,
	lastParameterLine int) bool {

	if funcType.Results == nil || len(funcType.Params.List) == 0 {
		return true
	}

	parameters := functionParameterParts(funcType.Params.List)
	lastParameterColumn := fset.Position(parameters[len(parameters)-1].position).Column
	lastLineWidth := lastParameterColumn - 1 + len(signature.parameters[len(signature.parameters)-1]) + len(signature.suffix)
	if lastLineWidth > maxFunctionDeclarationColumns {
		return true
	}

	return fset.Position(funcType.Results.Pos()).Line == lastParameterLine &&
		fset.Position(funcType.End()).Line == lastParameterLine
}

func buildFunctionSignature(function *ast.FuncDecl) functionSignature {
	prefix := "func "
	if function.Recv != nil {
		receiver := renderFields(function.Recv.List)
		prefix += "(" + strings.Join(receiver, ", ") + ") "
	}

	prefix += function.Name.Name
	if function.Type.TypeParams != nil {
		typeParameters := renderFields(function.Type.TypeParams.List)
		prefix += "[" + strings.Join(typeParameters, ", ") + "]"
	}
	prefix += "("

	parameters := renderParameterParts(function.Type.Params.List)
	suffix := renderSignatureSuffix(function.Type, function.Body != nil)

	return functionSignature{
		prefixWidth:       len(prefix),
		continuationWidth: 1,
		parameters:        parameters,
		suffix:            suffix,
	}
}

func buildTypeSignature(fset *token.FileSet, declaration *ast.GenDecl, typeSpec *ast.TypeSpec,
	funcType *ast.FuncType) functionSignature {

	continuationWidth := 1
	if declaration.Lparen.IsValid() {
		continuationWidth = fset.Position(typeSpec.Name.Pos()).Column
	}

	parameters := renderParameterParts(funcType.Params.List)
	suffix := renderSignatureSuffix(funcType, false)

	return functionSignature{
		prefixWidth:       fset.Position(funcType.Params.Opening).Column,
		continuationWidth: continuationWidth,
		parameters:        parameters,
		suffix:            suffix,
	}
}

func renderParameterParts(fields []*ast.Field) []string {
	rendered := make([]string, 0, len(fields))

	for _, field := range fields {
		typeName := types.ExprString(field.Type)
		if len(field.Names) == 0 {
			rendered = append(rendered, typeName)
			continue
		}

		for index, name := range field.Names {
			part := name.Name
			if index == len(field.Names)-1 {
				part += " " + typeName
			}
			rendered = append(rendered, part)
		}
	}

	return rendered
}

func functionParameterParts(fields []*ast.Field) []functionParameterPart {
	parts := make([]functionParameterPart, 0, len(fields))

	for _, field := range fields {
		if len(field.Names) == 0 {
			parts = append(parts, functionParameterPart{
				position: field.Type.Pos(),
				end:      field.Type.End()})
			continue
		}

		for index, name := range field.Names {
			end := name.End()
			if index == len(field.Names)-1 {
				end = field.Type.End()
			}
			parts = append(parts, functionParameterPart{
				position: name.Pos(),
				end:      end})
		}
	}

	return parts
}

func renderSignatureSuffix(funcType *ast.FuncType, hasBody bool) string {
	suffix := ")"
	if funcType.Results != nil {
		results := renderFields(funcType.Results.List)

		if len(funcType.Results.List) == 1 && len(funcType.Results.List[0].Names) == 0 {
			suffix += " " + results[0]
		} else {
			suffix += " (" + strings.Join(results, ", ") + ")"
		}
	}
	if hasBody {
		suffix += " {"
	}

	return suffix
}

func renderFields(fields []*ast.Field) []string {
	rendered := make([]string, 0, len(fields))

	for _, field := range fields {
		typeName := types.ExprString(field.Type)
		if len(field.Names) == 0 {
			rendered = append(rendered, typeName)
			continue
		}

		names := make([]string, 0, len(field.Names))
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
		rendered = append(rendered, strings.Join(names, ", ")+" "+typeName)
	}

	return rendered
}

func hasSignatureComments(file *ast.File, start, end token.Pos) bool {
	for _, comment := range file.Comments {
		if comment.Pos() > start && comment.Pos() < end {
			return true
		}
	}

	return false
}

func signatureEnd(funcType *ast.FuncType, body *ast.BlockStmt) token.Pos {
	if body != nil {
		return body.Lbrace
	}

	return funcType.End() - 1
}

func hasMultilineSignatureField(fset *token.FileSet, receiver, typeParams *ast.FieldList, funcType *ast.FuncType) bool {
	fieldLists := []*ast.FieldList{receiver, typeParams, funcType.Params, funcType.Results}

	for _, fields := range fieldLists {
		if fields == nil {
			continue
		}
		for _, field := range fields.List {
			if fset.Position(field.Type.Pos()).Line != fset.Position(field.Type.End()).Line {
				return true
			}
		}
	}

	return false
}
