package collections

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"

	"github.com/retroenv/retrogolib/set"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const retrogolibSetImportPath = "github.com/retroenv/retrogolib/set"

// SetProjectionRule detects the legacy Set.ToSlice projection.
type SetProjectionRule struct{}

// NewSetProjectionRule creates a new SetProjectionRule.
func NewSetProjectionRule() *SetProjectionRule {
	return &SetProjectionRule{}
}

// Name returns the rule name.
func (r *SetProjectionRule) Name() string {
	return "collections-set-projection"
}

// Description returns the rule description.
func (r *SetProjectionRule) Description() string {
	return "Use set.Sorted or set.SortedFunc instead of Set.ToSlice"
}

// Severity returns the default severity.
func (r *SetProjectionRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *SetProjectionRule) Category() string {
	return api.CategoryCollections
}

// Check analyzes a file for legacy set projections.
func (r *SetProjectionRule) Check(fset *token.FileSet, file *ast.File) []violation.Violation {
	if !importsRetrogolibSet(file) {
		return nil
	}

	var violations []violation.Violation
	source := newSetProjectionSource(fset, file)

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || len(call.Args) != 0 || !isToSliceCall(call) {
			return true
		}
		selector := call.Fun.(*ast.SelectorExpr)
		if !source.isReceiver(selector.X, set.New[types.Object]()) {
			return true
		}

		violations = append(violations, violation.Violation{
			Rule:     r.Name(),
			Message:  "Use set.Sorted or set.SortedFunc instead of Set.ToSlice",
			Position: fset.Position(call.Pos()),
			Severity: r.Severity(),
		})

		return true
	})

	return violations
}

type setProjectionSource struct {
	declarations map[types.Object]ast.Node
	info         *types.Info
}

func newSetProjectionSource(fset *token.FileSet, file *ast.File) *setProjectionSource {
	source := &setProjectionSource{
		declarations: make(map[types.Object]ast.Node),
		info: &types.Info{
			Defs: make(map[*ast.Ident]types.Object),
			Uses: make(map[*ast.Ident]types.Object),
		},
	}
	// Resolve lexical bindings without loading dependencies. Missing imports and
	// other files can leave types unknown; only proven set receivers are reported.
	config := types.Config{Error: func(error) {}}
	_, _ = config.Check(file.Name.Name, fset, []*ast.File{file}, source.info)
	ast.Inspect(file, func(node ast.Node) bool {
		var names []*ast.Ident
		switch declaration := node.(type) {
		case *ast.Field:
			names = declaration.Names
		case *ast.TypeSpec:
			names = []*ast.Ident{declaration.Name}
		case *ast.FuncDecl:
			names = []*ast.Ident{declaration.Name}
		case *ast.ValueSpec:
			names = declaration.Names
		case *ast.AssignStmt:
			for _, left := range declaration.Lhs {
				if name, ok := left.(*ast.Ident); ok {
					names = append(names, name)
				}
			}
		}
		for _, name := range names {
			if object := source.info.Defs[name]; object != nil {
				source.declarations[object] = node
			}
		}
		return true
	})
	return source
}

// isReceiver requires a set declaration or constructor in this file.
// An import alone does not establish the receiver type.
func (source *setProjectionSource) isReceiver(expr ast.Expr, seen set.Set[types.Object]) bool {
	switch expr := expr.(type) {
	case *ast.ParenExpr:
		return source.isReceiver(expr.X, seen)
	case *ast.StarExpr:
		return source.isReceiver(expr.X, seen)
	case *ast.IndexExpr:
		return source.isReceiver(expr.X, seen)
	case *ast.CompositeLit:
		return source.isReceiver(expr.Type, seen)
	case *ast.SelectorExpr:
		if expr.Sel.Name == "Set" && source.isPackage(expr.X) {
			return true
		}
		return source.isReceiver(expr.Sel, seen)
	case *ast.CallExpr:
		return source.isConstructor(expr, seen)
	case *ast.Ident:
		object := source.info.ObjectOf(expr)
		if object == nil || seen.Contains(object) {
			return false
		}
		seen.Add(object)
		return source.isDeclaration(object, seen)
	default:
		return false
	}
}

func (source *setProjectionSource) isDeclaration(object types.Object, seen set.Set[types.Object]) bool {
	switch declaration := source.declarations[object].(type) {
	case *ast.Field:
		return source.isReceiver(declaration.Type, seen)
	case *ast.TypeSpec:
		return declaration.Assign.IsValid() && source.isReceiver(declaration.Type, seen)
	case *ast.FuncDecl:
		results := declaration.Type.Results
		return results != nil && len(results.List) == 1 && source.isReceiver(results.List[0].Type, seen)
	case *ast.ValueSpec:
		if declaration.Type != nil {
			return source.isReceiver(declaration.Type, seen)
		}
		for index, name := range declaration.Names {
			if source.info.ObjectOf(name) == object && index < len(declaration.Values) {
				return source.isReceiver(declaration.Values[index], seen)
			}
		}
	case *ast.AssignStmt:
		return source.isAssignment(object, declaration, seen)
	}
	return false
}

func (source *setProjectionSource) isConstructor(call *ast.CallExpr, seen set.Set[types.Object]) bool {
	function := call.Fun
	if indexed, ok := function.(*ast.IndexExpr); ok {
		function = indexed.X
	}
	if identifier, ok := function.(*ast.Ident); ok {
		if identifier.Name == "make" && source.info.ObjectOf(identifier) == types.Universe.Lookup("make") &&
			len(call.Args) != 0 {

			return source.isReceiver(call.Args[0], seen)
		}
		return source.isReceiver(identifier, seen)
	}
	selector, ok := function.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	if source.isPackage(selector.X) {
		return selector.Sel.Name == "New" || selector.Sel.Name == "NewFromSlice" || selector.Sel.Name == "Set"
	}
	if source.isReceiver(selector.Sel, seen) {
		return true
	}
	switch selector.Sel.Name {
	case "Copy", "Union", "Intersection", "Difference", "SymmetricDifference":
		return source.isReceiver(selector.X, seen)
	default:
		return false
	}
}

func (source *setProjectionSource) isPackage(expr ast.Expr) bool {
	identifier, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	name, ok := source.info.ObjectOf(identifier).(*types.PkgName)
	return ok && name.Imported().Path() == retrogolibSetImportPath
}

func (source *setProjectionSource) isAssignment(object types.Object, assignment *ast.AssignStmt,
	seen set.Set[types.Object]) bool {

	for index, left := range assignment.Lhs {
		name, ok := left.(*ast.Ident)
		if ok && source.info.ObjectOf(name) == object && index < len(assignment.Rhs) {
			return source.isReceiver(assignment.Rhs[index], seen)
		}
	}
	return false
}

func importsRetrogolibSet(file *ast.File) bool {
	for _, importSpec := range file.Imports {
		path, err := strconv.Unquote(importSpec.Path.Value)
		if err == nil && path == retrogolibSetImportPath {
			return true
		}
	}

	return false
}

func isToSliceCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "ToSlice"
}
