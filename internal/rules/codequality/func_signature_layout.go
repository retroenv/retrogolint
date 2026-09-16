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

type functionSignature struct {
	prefix     string
	parameters []string
	suffix     string
}

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
	return "Function signatures should use the available 120 columns and wrap between complete parameters"
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
			Message:  function.Name.Name + ": function signature should use available 120 columns and wrap between complete parameters",
			Position: fset.Position(function.Name.Pos()),
			Severity: r.Severity(),
		})
	}

	return violations
}

func (signature functionSignature) hasValidLayout(fset *token.FileSet, function *ast.FuncDecl) bool {
	startLine := fset.Position(function.Type.Func).Line
	endLine := fset.Position(functionSignatureEnd(function)).Line
	fullWidth := len(signature.prefix) + len(strings.Join(signature.parameters, ", ")) + len(signature.suffix)

	if fullWidth <= maxFunctionDeclarationColumns {
		return startLine == endLine
	}

	if startLine == endLine || fset.Position(function.Type.Params.Opening).Line != startLine {
		return false
	}

	expectedFirstLineParameters := signature.firstLineParameterCount()
	actualFirstLineParameters := firstLineParameterCount(fset, function)
	if actualFirstLineParameters != expectedFirstLineParameters {
		return false
	}

	parameters := function.Type.Params.List
	for _, parameter := range parameters {
		if fset.Position(parameter.Pos()).Line != fset.Position(parameter.End()).Line {
			return false
		}

		if fset.Position(parameter.End()-1).Column > maxFunctionDeclarationColumns {
			return false
		}
	}

	lastParameterLine := fset.Position(function.Type.Params.Opening).Line
	if len(parameters) > 0 {
		lastParameterLine = fset.Position(parameters[len(parameters)-1].End()).Line
	}
	if fset.Position(function.Type.Params.Closing).Line != lastParameterLine {
		return false
	}

	return signature.hasValidResultLayout(fset, function, lastParameterLine)
}

func (signature functionSignature) firstLineParameterCount() int {
	line := signature.prefix

	for index, parameter := range signature.parameters {
		candidate := line + parameter
		if index < len(signature.parameters)-1 {
			candidate += ","
		} else {
			candidate += signature.suffix
		}
		if len(candidate) > maxFunctionDeclarationColumns {
			return index
		}

		line += parameter
		if index < len(signature.parameters)-1 {
			line += ", "
		}
	}

	return len(signature.parameters)
}

func (signature functionSignature) hasValidResultLayout(fset *token.FileSet, function *ast.FuncDecl,
	lastParameterLine int) bool {

	if function.Body != nil && fset.Position(function.Body.Lbrace).Column > maxFunctionDeclarationColumns {
		return false
	}

	if function.Type.Results == nil || len(function.Type.Params.List) == 0 {
		return true
	}

	lastParameter := function.Type.Params.List[len(function.Type.Params.List)-1]
	lastParameterColumn := fset.Position(lastParameter.Pos()).Column
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

	parameters := renderFields(function.Type.Params.List)
	suffix := renderFunctionSuffix(function)

	return functionSignature{
		prefix:     prefix,
		parameters: parameters,
		suffix:     suffix,
	}
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

func firstLineParameterCount(fset *token.FileSet, function *ast.FuncDecl) int {
	openingLine := fset.Position(function.Type.Params.Opening).Line
	count := 0

	for _, parameter := range function.Type.Params.List {
		if fset.Position(parameter.Pos()).Line != openingLine {
			break
		}
		count++
	}

	return count
}
