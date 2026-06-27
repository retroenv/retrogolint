package loggingfields

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

// LoggingSpecializedFieldRule ensures log calls use specialized field helpers instead of casting.
type LoggingSpecializedFieldRule struct{}

// NewLoggingSpecializedFieldRule creates a new LoggingSpecializedFieldRule.
func NewLoggingSpecializedFieldRule() *LoggingSpecializedFieldRule {
	return &LoggingSpecializedFieldRule{}
}

// Name returns the rule name.
func (r *LoggingSpecializedFieldRule) Name() string {
	return "logging-specialized-field"
}

// Description returns the rule description.
func (r *LoggingSpecializedFieldRule) Description() string {
	return "Prefer specialized log field helpers (e.g., log.Uint16) instead of casting to generic integers"
}

// Severity returns the default severity.
func (r *LoggingSpecializedFieldRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *LoggingSpecializedFieldRule) Category() string {
	return api.CategoryLogging
}

// Check analyzes a file for violations.
func (r *LoggingSpecializedFieldRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	var violations []violation.Violation

	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	conf := types.Config{
		Importer: importer.Default(),
		Error:    func(_ error) {}, // Ignore type errors
	}

	_, _ = conf.Check("", fset, []*ast.File{file}, info)

	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		typedIdents := r.collectTypedIdents(fn)

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			logCall, ok := api.GetLogCallInfo(call)
			if !ok {
				return true
			}

			for i := logCall.FieldStartIndex; i < len(call.Args); i++ {
				fieldCall, ok := call.Args[i].(*ast.CallExpr)
				if !ok || !api.IsLogFieldCall(fieldCall) {
					continue
				}

				if violation := r.checkField(fset, fieldCall, typedIdents, info); violation != nil {
					violations = append(violations, *violation)
				}
			}

			return true
		})
	}

	return violations
}

// extractBasicType extracts the type name from a types.Type.
func (r *LoggingSpecializedFieldRule) extractBasicType(t types.Type) string {
	if named, ok := t.(*types.Named); ok {
		t = named.Underlying()
	}

	if basic, ok := t.(*types.Basic); ok {
		return basic.Name()
	}

	return ""
}

var typeToField = map[string]string{
	"int":    "Int",
	"int64":  "Int64",
	"int32":  "Int32",
	"int16":  "Int16",
	"int8":   "Int8",
	"uint":   "Uint",
	"uint64": "Uint64",
	"uint32": "Uint32",
	"uint16": "Uint16",
	"uint8":  "Uint8",
}

var conversionTargets = map[string]string{
	"int":  "Int",
	"uint": "Uint",
}

func (r *LoggingSpecializedFieldRule) resolveOperandType(info *types.Info, operand ast.Expr, typedIdents map[string]string) string {
	if info != nil {
		if tv, ok := info.Types[operand]; ok {
			if t := r.extractBasicType(tv.Type); t != "" {
				return t
			}
		}
	}

	ident, ok := operand.(*ast.Ident)
	if !ok {
		return ""
	}

	return typedIdents[ident.Name]
}

func (r *LoggingSpecializedFieldRule) checkField(fset *token.FileSet, call *ast.CallExpr, typedIdents map[string]string, info *types.Info) *violation.Violation {
	if len(call.Args) < 2 {
		return nil
	}

	value := unwrapParen(call.Args[1])
	convCall, ok := value.(*ast.CallExpr)
	if !ok || len(convCall.Args) != 1 {
		return nil
	}

	convName, ok := r.extractConversion(conversionTargets, convCall.Fun)
	if !ok {
		return nil
	}

	operand := unwrapParen(convCall.Args[0])
	operandType := r.resolveOperandType(info, operand, typedIdents)

	if operandType == "" {
		return nil
	}

	specialized, ok := typeToField[operandType]
	if !ok {
		return nil
	}

	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}

	if sel.Sel.Name == specialized {
		return nil
	}

	if conversionTargets[convName] != sel.Sel.Name {
		return nil
	}

	pos := fset.Position(call.Pos())

	return &violation.Violation{
		Rule:     r.Name(),
		Message:  fmt.Sprintf("Use log.%s for %s values instead of casting to %s", specialized, operandType, convName),
		Position: pos,
		Severity: r.Severity(),
	}
}

func (r *LoggingSpecializedFieldRule) extractConversion(targets map[string]string, fun ast.Expr) (string, bool) {
	ident, ok := fun.(*ast.Ident)
	if !ok {
		return "", false
	}

	if _, ok := targets[ident.Name]; !ok {
		return "", false
	}

	return ident.Name, true
}

func (r *LoggingSpecializedFieldRule) collectVarDecl(decl *ast.GenDecl, typed map[string]string) {
	if decl.Tok != token.VAR {
		return
	}

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok || valueSpec.Type == nil {
			continue
		}

		r.collectValueSpecTypes(valueSpec, typed)
	}
}

func (r *LoggingSpecializedFieldRule) collectTypedIdents(fn *ast.FuncDecl) map[string]string {
	typed := make(map[string]string)

	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			r.collectFieldTypes(field, typed)
		}
	}

	if fn.Recv != nil {
		for _, field := range fn.Recv.List {
			r.collectFieldTypes(field, typed)
		}
	}

	if fn.Body == nil {
		return typed
	}

	ast.Inspect(fn.Body, func(node ast.Node) bool {
		switch stmt := node.(type) {
		case *ast.DeclStmt:
			if decl, ok := stmt.Decl.(*ast.GenDecl); ok {
				r.collectVarDecl(decl, typed)
			}
		case *ast.GenDecl:
			r.collectVarDecl(stmt, typed)
		}

		return true
	})

	return typed
}

func (r *LoggingSpecializedFieldRule) collectFieldTypes(field *ast.Field, typed map[string]string) {
	typ := unwrapTypeName(field.Type)
	if typ == "" {
		return
	}

	for _, name := range field.Names {
		if name == nil {
			continue
		}
		typed[name.Name] = typ
	}
}

func (r *LoggingSpecializedFieldRule) collectValueSpecTypes(spec *ast.ValueSpec, typed map[string]string) {
	typ := unwrapTypeName(spec.Type)
	if typ == "" {
		return
	}

	for _, name := range spec.Names {
		if name == nil {
			continue
		}
		typed[name.Name] = typ
	}
}

func unwrapTypeName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func unwrapParen(expr ast.Expr) ast.Expr {
	for {
		paren, ok := expr.(*ast.ParenExpr)
		if !ok {
			break
		}
		expr = paren.X
	}
	return expr
}
