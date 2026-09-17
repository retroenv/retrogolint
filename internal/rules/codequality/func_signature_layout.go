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

// Check analyzes function and method signature layout.
func (r *FuncSignatureLayoutRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || hasSignatureComments(file, function) {
			continue
		}

		signature := buildFunctionSignature(function)
		if signature.hasValidLayout(fset, function) {
			continue
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  function.Name.Name + ": function signature should use available 120 columns and wrap after parameter commas",
			Position: fset.Position(function.Name.Pos()),
			Severity: r.Severity(),
		})
	}

	return violations
}

type functionSignature struct {
	prefix     string
	parameters []string
	suffix     string
}

type functionParameterPart struct {
	position token.Pos
	end      token.Pos
}

func (signature functionSignature) hasValidLayout(fset *token.FileSet, function *ast.FuncDecl) bool {
	if hasMultilineSignatureField(fset, function) {
		return true
	}

	startLine := fset.Position(function.Type.Func).Line
	endLine := fset.Position(functionSignatureEnd(function)).Line
	fullWidth := len(signature.prefix) + len(strings.Join(signature.parameters, ", ")) + len(signature.suffix)

	if fullWidth <= maxFunctionDeclarationColumns {
		return startLine == endLine
	}

	if startLine == endLine {
		return signature.requiresIndivisibleOverrun()
	}
	if fset.Position(function.Type.Params.Opening).Line != startLine {
		return false
	}

	parameters := functionParameterParts(function.Type.Params.List)
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

	lastParameterLine := fset.Position(function.Type.Params.Opening).Line
	if len(parameters) > 0 {
		lastParameterLine = fset.Position(parameters[len(parameters)-1].end).Line
	}
	if fset.Position(function.Type.Params.Closing).Line != lastParameterLine {
		return false
	}

	return signature.hasValidResultLayout(fset, function, lastParameterLine)
}

func (signature functionSignature) requiresIndivisibleOverrun() bool {
	if len(signature.parameters) == 0 {
		return true
	}

	lastParameter := signature.parameters[len(signature.parameters)-1]
	return 1+len(lastParameter)+len(signature.suffix) > maxFunctionDeclarationColumns
}

func (signature functionSignature) parameterLineIndexes() []int {
	lines := make([]int, len(signature.parameters))
	lineIndex := 0
	lineWidth := len(signature.prefix)
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
			lineWidth = 1
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
	if 1+len(parameter)+len(signature.suffix) <= maxFunctionDeclarationColumns {
		return signature.suffix
	}

	return ")"
}

func (signature functionSignature) hasValidResultLayout(fset *token.FileSet, function *ast.FuncDecl,
	lastParameterLine int) bool {

	if function.Type.Results == nil || len(function.Type.Params.List) == 0 {
		return true
	}

	parameters := functionParameterParts(function.Type.Params.List)
	lastParameterColumn := fset.Position(parameters[len(parameters)-1].position).Column
	lastLineWidth := lastParameterColumn - 1 + len(signature.parameters[len(signature.parameters)-1]) + len(signature.suffix)
	if lastLineWidth > maxFunctionDeclarationColumns {
		return true
	}

	return fset.Position(function.Type.Results.Pos()).Line == lastParameterLine &&
		fset.Position(function.Type.End()).Line == lastParameterLine
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
	suffix := renderFunctionSuffix(function)

	return functionSignature{
		prefix:     prefix,
		parameters: parameters,
		suffix:     suffix,
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

func renderFunctionSuffix(function *ast.FuncDecl) string {
	suffix := ")"
	if function.Type.Results != nil {
		results := renderFields(function.Type.Results.List)

		if len(function.Type.Results.List) == 1 && len(function.Type.Results.List[0].Names) == 0 {
			suffix += " " + results[0]
		} else {
			suffix += " (" + strings.Join(results, ", ") + ")"
		}
	}
	if function.Body != nil {
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

func hasSignatureComments(file *ast.File, function *ast.FuncDecl) bool {
	end := functionSignatureEnd(function)

	for _, comment := range file.Comments {
		if comment.Pos() > function.Type.Func && comment.Pos() < end {
			return true
		}
	}

	return false
}

func functionSignatureEnd(function *ast.FuncDecl) token.Pos {
	if function.Body != nil {
		return function.Body.Lbrace
	}

	return function.Type.End() - 1
}

func hasMultilineSignatureField(fset *token.FileSet, function *ast.FuncDecl) bool {
	fieldLists := []*ast.FieldList{function.Recv, function.Type.TypeParams, function.Type.Params, function.Type.Results}

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
