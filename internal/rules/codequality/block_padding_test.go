package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

type blockPaddingTest struct {
	name         string
	code         string
	wantLines    []int
	wantMessages []string
}

const (
	leadingMessage   = "block should not start with an empty line"
	trailingMessage  = "block should not end with an empty line"
	conditionMessage = "multiline if condition should be followed by an empty line or a comment"
	signatureMessage = "multiline function signature should be followed by an empty line or a comment"
)

func TestBlockPaddingRule_LeadingAndTrailing(t *testing.T) {
	tests := []blockPaddingTest{
		{
			name: "compact blocks",
			code: `package test
func run(a int) int {
	if a > 0 {
		a++
	}
	return a
}
`,
		},
		{
			name: "empty line at start of function body",
			code: `package test
func run(a int) int {

	a++
	return a
}
`,
			wantLines:    []int{2},
			wantMessages: []string{leadingMessage},
		},
		{
			name: "empty line at end of if body",
			code: `package test
func run(a int) int {
	if a > 0 {
		a++

	}
	return a
}
`,
			wantLines:    []int{6},
			wantMessages: []string{trailingMessage},
		},
		{
			name: "empty line at start and end of block",
			code: `package test
func run(a int) {
	for a < 10 {

		a++

	}
}
`,
			wantLines:    []int{3, 7},
			wantMessages: []string{leadingMessage, trailingMessage},
		},
		{
			name: "empty line at start of switch body",
			code: `package test
func run(a int) {
	switch a {

	case 1:
		a++
	}
}
`,
			wantLines:    []int{3},
			wantMessages: []string{leadingMessage},
		},
		{
			name: "empty line inside block is allowed",
			code: `package test
func run(a int) int {
	a++

	return a
}
`,
		},
		{
			name: "empty block",
			code: `package test
func run() {

}
`,
		},
		{
			name: "single line block",
			code: `package test
func run(a int) int { return a }
`,
		},
	}

	runBlockPaddingTests(t, tests)
}

func TestBlockPaddingRule_Comments(t *testing.T) {
	tests := []blockPaddingTest{
		{
			name: "empty line after leading comment",
			code: `package test
func run(a int) int {
	// increment

	a++
	return a
}
`,
		},
		{
			name: "empty line before leading comment",
			code: `package test
func run(a int) int {

	// increment
	a++
	return a
}
`,
			wantLines:    []int{2},
			wantMessages: []string{leadingMessage},
		},
		{
			name: "comment at end of block",
			code: `package test
func run(a int) int {
	a++
	// done
}
`,
		},
		{
			name: "empty line before trailing comment",
			code: `package test
func run(a int) {
	a++

	// done
}
`,
		},
		{
			name: "empty line after trailing comment",
			code: `package test
func run(a int) {
	a++
	// done

}
`,
			wantLines:    []int{6},
			wantMessages: []string{trailingMessage},
		},
		{
			name: "comment after opening brace",
			code: `package test
func run(a int) int { // run
	a++
	return a
}
`,
		},
		{
			name: "empty line after comment on opening brace line",
			code: `package test
func run(a int) int { // run

	a++
	return a
}
`,
			wantLines:    []int{2},
			wantMessages: []string{leadingMessage},
		},
		{
			name: "multiline comment on opening brace line",
			code: `package test
func run(a int) int { /* run
	the code */

	a++
	return a
}
`,
		},
	}

	runBlockPaddingTests(t, tests)
}

func TestBlockPaddingRule_MultilineHeaders(t *testing.T) {
	tests := []blockPaddingTest{
		{
			name: "multiline if condition directly followed by statement",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 {
		a++
	}
}
`,
			wantLines:    []int{4},
			wantMessages: []string{conditionMessage},
		},
		{
			name: "multiline if condition followed by empty line",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 {

		a++
	}
}
`,
		},
		{
			name: "multiline if condition followed by comment",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 {
		// increment
		a++
	}
}
`,
		},
		{
			name: "multiline if condition with trailing comment on brace line",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 { // check
		a++
	}
}
`,
			wantLines:    []int{4},
			wantMessages: []string{conditionMessage},
		},
		{
			name: "multiline else if condition",
			code: `package test
func run(a, b int) {
	if a > 5 {
		a++
	} else if a > 0 &&
		b > 0 {
		a--
	}
}
`,
			wantLines:    []int{6},
			wantMessages: []string{conditionMessage},
		},
		{
			name: "multiline if init with single line condition",
			code: `package test
func run(a int) {
	if b := max(
		a, 1); b > 0 {
		a++
	}
}
`,
		},
		{
			name: "single line if condition",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
	}
}
`,
		},
		{
			name: "empty body with multiline condition",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 {
	}
}
`,
		},
		{
			name: "multiline function signature",
			code: `package test
func run(a int,
	b int) int {
	return a + b
}
`,
			wantLines:    []int{3},
			wantMessages: []string{signatureMessage},
		},
		{
			name: "multiline function signature followed by empty line",
			code: `package test
func run(a int,
	b int) int {

	return a + b
}
`,
		},
		{
			name: "multiline function literal signature",
			code: `package test
func run() func(a, b int) int {
	return func(a int,
		b int) int {
		return a + b
	}
}
`,
			wantLines:    []int{4},
			wantMessages: []string{signatureMessage},
		},
		{
			name: "function declaration without body",
			code: `package test
func run(a int,
	b int) int
`,
		},
		{
			name: "multiline function call is not a signature",
			code: `package test
func run() int {
	return add(1,
		2)
}
`,
		},
		{
			name: "empty line at end of body with multiline condition",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 {

		a++

	}
}
`,
			wantLines:    []int{8},
			wantMessages: []string{trailingMessage},
		},
		{
			name: "comment at end of body with multiline condition",
			code: `package test
func run(a, b int) {
	if a > 0 &&
		b > 0 {

		a++
		// done
	}
}
`,
		},
	}

	runBlockPaddingTests(t, tests)
}

func TestBlockPaddingRule_Metadata(t *testing.T) {
	rule := NewBlockPaddingRule()

	assert.Equal(t, "codequality-block-padding", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
	assert.NotEmpty(t, rule.Description())
}

func runBlockPaddingTests(t *testing.T, tests []blockPaddingTest) {
	t.Helper()
	rule := NewBlockPaddingRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			lines := make([]int, 0, len(violations))
			messages := make([]string, 0, len(violations))
			for _, v := range violations {
				lines = append(lines, v.Position.Line)
				messages = append(messages, v.Message)
			}
			assert.Len(t, lines, len(tt.wantLines))
			for index, line := range tt.wantLines {
				assert.Equal(t, line, lines[index])
				assert.Equal(t, tt.wantMessages[index], messages[index])
			}
		})
	}
}
