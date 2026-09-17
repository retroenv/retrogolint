package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestFuncSignatureLayoutRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "short signature on one line",
			code: `package test
func Add(left, right int) int { return left + right }
`,
		},
		{
			name: "short signature split across lines",
			code: `package test
func Add(
	left, right int,
) int { return left + right }
`,
			wantViolations: 1,
		},
		{
			name: "reported declaration regression",
			code: `package test
func OptimizeUnreferencedLabelsProtecting(
	sequence machineopt.Sequence[ast.Node],
	protected set.Set[string],
) int {
	return 0
}
`,
			wantViolations: 1,
		},
		{
			name: "reported declaration on one line",
			code: `package test
func OptimizeUnreferencedLabelsProtecting(sequence machineopt.Sequence[ast.Node], protected set.Set[string]) int {
	return 0
}
`,
		},
		{
			name: "long signature wrapped after parameters that fit",
			code: `package test
func BuildRuntimeSnapshot(ctx context.Context, dependencies RuntimeDependencies, loaded LoadedConfig,
	options ExtremelyLongOptions) (*RuntimeSnapshot, error) {

	return nil, nil
}
`,
		},
		{
			name: "continuation line packs parameters that fit",
			code: `package test
func OptimizeAdjacentValueHomeCoalesceAtBase(logger *log.Logger, sequence machineopt.Sequence[ast.Node],
	b MemoryHomePolicy, codeBase uint64) int {

	return 0
}
`,
		},
		{
			name: "continuation line leaves avoidable space",
			code: `package test
func OptimizeAdjacentValueHomeCoalesceAtBase(logger *log.Logger, sequence machineopt.Sequence[ast.Node],
	b MemoryHomePolicy,
	codeBase uint64) int {

	return 0
}
`,
			wantViolations: 1,
		},
		{
			name: "result tuple stays with final parameter",
			code: `package test
func findDeadTransferBranch(nodes []ast.Node, bridge optmanager.Bridge,
	includeDead, includeDelayed bool) (int, *deadTransferBranchAlias, bool) {

	return 0, nil, false
}
`,
		},
		{
			name: "long signature moves all parameters",
			code: `package test
func BuildRuntimeSnapshot(
	ctx context.Context,
	dependencies RuntimeDependencies,
	loaded LoadedConfig,
	options ExtremelyLongOptions) (*RuntimeSnapshot, error) {

	return nil, nil
}
`,
			wantViolations: 1,
		},
		{
			name: "long signature has standalone closing parenthesis",
			code: `package test
func BuildRuntimeSnapshot(ctx context.Context, dependencies RuntimeDependencies, loaded LoadedConfig,
	options ExtremelyLongOptions,
) (*RuntimeSnapshot, error) {

	return nil, nil
}
`,
			wantViolations: 1,
		},
		{
			name: "long signature moves result unnecessarily",
			code: `package test
func BuildRuntimeSnapshot(ctx context.Context, dependencies RuntimeDependencies, loaded LoadedConfig,
	options ExtremelyLongOptions) (
	*RuntimeSnapshot,
	error) {

	return nil, nil
}
`,
			wantViolations: 1,
		},
		{
			name: "long signature remains on one line",
			code: `package test
func BuildRuntimeSnapshot(ctx context.Context, dependencies RuntimeDependencies, loaded LoadedConfig, options ExtremelyLongOptions) (*RuntimeSnapshot, error) { return nil, nil }
`,
			wantViolations: 1,
		},
		{
			name: "generic function wraps complete parameter",
			code: `package test
func TransformValues[Input any, Output any](ctx context.Context, values []Input, transform func(Input) Output,
	options ExtremelyLongTransformationOptions) ([]Output, error) {

	return nil, nil
}
`,
		},
		{
			name: "method signature uses receiver width",
			code: `package test
type Transformer struct{}
func (transformer *Transformer) TransformValues(ctx context.Context, values []Input,
	options ExtremelyLongTransformationOptions) ([]Output, error) {

	return nil, nil
}
`,
		},
		{
			name: "indivisible function result may exceed line width",
			code: `package test
func BuildCallback(values []Value) func(int, machineopt.InstructionView, machineopt.ResourceSet, machineopt.ResourceSet) (machineopt.ResourceSet, machineopt.ResourceSet) {

	return nil
}
`,
		},
		{
			name: "indivisible function parameter may exceed line width",
			code: `package test
func Apply(values []Value,
	build func(machineopt.InstructionWindow[Node], analysis.ResourceFacts) (machineopt.SequenceReplacement[Node], bool)) bool {

			return build != nil
}
`,
		},
		{
			name: "anonymous struct result keeps intrinsic layout",
			code: `package test
func Cases() []struct {
	name string
	want bool
} {

	return nil
}
`,
		},
		{
			name: "signature comments are ignored",
			code: `package test
func Add(
	// Keep this parameter explanation.
	left int,
	right int,
) int { return left + right }
`,
		},
	}

	rule := NewFuncSignatureLayoutRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestFuncSignatureLayoutRule_Metadata(t *testing.T) {
	rule := NewFuncSignatureLayoutRule()

	assert.Equal(t, "codequality-func-signature-layout", rule.Name())
	assert.Equal(t, "Function signatures should use the available 120 columns and wrap between complete parameters", rule.Description())
	assert.Equal(t, "codequality", rule.Category())
}
