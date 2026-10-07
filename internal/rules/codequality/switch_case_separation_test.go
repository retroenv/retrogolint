package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

type switchCaseSeparationTest struct {
	name      string
	code      string
	wantLines []int
}

func TestSwitchCaseSeparationRule_Check(t *testing.T) {
	tests := []switchCaseSeparationTest{
		{
			name: "multiline case body directly before next case",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fdsOverride |= 0x80
		return false
	case 0x58:
		fdsModDepth = nextByte(c)
		return false
	}
	return true
}
`,
			wantLines: []int{8},
		},
		{
			name: "empty line separates cases",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fdsOverride |= 0x80
		return false

	case 0x58:
		fdsModDepth = nextByte(c)
		return false
	}
	return true
}
`,
		},
		{
			name: "single statement bodies",
			code: `package test
func expansionCommand(op byte) bool {
	switch op {
	case 0x57:
		return false
	case 0x58:
		return true
	}
	return true
}
`,
		},
		{
			name: "multiple violations in one switch",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fdsOverride |= 0x80
		return false
	case 0x58:
		fdsModDepth = nextByte(c)
		fdsOverride |= 0x40
	case 0x59:
		return true
	}
	return true
}
`,
			wantLines: []int{8, 11},
		},
		{
			name: "fallthrough body keeps next case attached",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fallthrough
	case 0x58:
		return false
	}
	return true
}
`,
		},
		{
			name: "default label directly follows multiline body",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fdsOverride |= 0x80
	default:
		return false
	}
	return true
}
`,
			wantLines: []int{7},
		},
		{
			name: "type switch with multiline body",
			code: `package test
func convert(value any) int {
	switch v := value.(type) {
	case int:
		result = 1
		result += v
	case string:
		return 0
	}
	return 0
}
`,
			wantLines: []int{7},
		},
		{
			name: "multiline body ending with block statement",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		if ok {
			fdsModSpeed = nextWord(c)
		}
	case 0x58:
		return false
	}
	return true
}
`,
			wantLines: []int{8},
		},
		{
			name: "multiline case expression with single statement body",
			code: `package test
func isPredeclaredType(name string) bool {
	switch name {
	case "any", "bool", "byte",
		"int", "int8", "int16":
		return true
	default:
		return false
	}
}
`,
		},
		{
			name: "empty case body",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
	case 0x58:
		return false
	}
	return true
}
`,
		},
		{
			name: "multiline body in last case before closing brace",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fdsOverride |= 0x80
	}
	return true
}
`,
		},
		{
			name: "single statement body on next line",
			code: `package test
func expansionCommand(op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
	case 0x58:
		return false
	}
	return true
}
`,
		},
		{
			name: "single statement case directly before multiline case",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		return false
	case 0x58:
		fdsModDepth = nextByte(c)
		return false
	}
	return true
}
`,
			wantLines: []int{6},
		},
		{
			name: "multiline cases separated from single statement cases",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x55:
		return true
	case 0x56:
		return false

	case 0x57:
		fdsModSpeed = nextWord(c)
		return false

	case 0x58:
		return true
	}
	return true
}
`,
		},
		{
			name: "empty case grouped with following multiline case",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x56:
		return true
	case 0x57:
	case 0x58:
		fdsModDepth = nextByte(c)
		return false
	}
	return true
}
`,
			wantLines: []int{6},
		},
		{
			name: "empty case after multiline body",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x56:
		fdsModDepth = nextByte(c)
		return false
	case 0x57:
	case 0x58:
		return true
	}
	return true
}
`,
			wantLines: []int{7},
		},
		{
			name: "fallthrough into multiline case",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fallthrough
	case 0x58:
		fdsModDepth = nextByte(c)
		return false
	}
	return true
}
`,
		},
	}

	runSwitchCaseSeparationTests(t, tests)
}

func TestSwitchCaseSeparationRule_Comments(t *testing.T) {
	tests := []switchCaseSeparationTest{
		{
			name: "comment line does not separate cases",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		return false
		// 0x58 adjusts the depth.
	case 0x58:
		return false
	}
	return true
}
`,
			wantLines: []int{8},
		},
		{
			name: "trailing comment on last body line",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		fdsOverride |= 0x80
		return false // shared line comment
	case 0x58:
		return false
	}
	return true
}
`,
			wantLines: []int{8},
		},
		{
			name: "comment after empty line",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		return false

	// 0x58 adjusts the depth.
	case 0x58:
		return false
	}
	return true
}
`,
		},
		{
			name: "comment before empty line",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		return false
		// The speed is in 1/16 steps.

	case 0x58:
		return false
	}
	return true
}
`,
		},
		{
			name: "trailing comment continues on next line",
			code: `package test
func expansionCommand(c, op byte) bool {
	switch op {
	case 0x57:
		fdsModSpeed = nextWord(c)
		return false // first comment line
		// second comment line
	case 0x58:
		return false
	}
	return true
}
`,
			wantLines: []int{8},
		},
	}

	runSwitchCaseSeparationTests(t, tests)
}

func TestSwitchCaseSeparationRule_Select(t *testing.T) {
	tests := []switchCaseSeparationTest{
		{
			name: "select statement with multiline body",
			code: `package test
func receive(values chan int, done chan struct{}) int {
	select {
	case value := <-values:
		total += value
		return total
	case <-done:
		return 0
	}
}
`,
			wantLines: []int{7},
		},
		{
			name: "select statement with separated cases",
			code: `package test
func receive(values chan int, done chan struct{}) int {
	select {
	case value := <-values:
		total += value
		return total

	case <-done:
		return 0
	}
}
`,
		},
	}

	runSwitchCaseSeparationTests(t, tests)
}

func TestSwitchCaseSeparationRule_Metadata(t *testing.T) {
	rule := NewSwitchCaseSeparationRule()

	assert.Equal(t, "codequality-switch-case-separation", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
}

func runSwitchCaseSeparationTests(t *testing.T, tests []switchCaseSeparationTest) {
	t.Helper()
	rule := NewSwitchCaseSeparationRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			lines := make([]int, 0, len(violations))
			for _, v := range violations {
				lines = append(lines, v.Position.Line)
			}
			assert.Len(t, lines, len(tt.wantLines))
			for index, line := range tt.wantLines {
				assert.Equal(t, line, lines[index])
			}
		})
	}
}
