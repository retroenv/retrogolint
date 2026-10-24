package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

type statementSeparationTest struct {
	name      string
	code      string
	wantLines []int
}

func TestStatementSeparationRule_ControlStatements(t *testing.T) {
	tests := []statementSeparationTest{
		{
			name: "statement directly after multiline if",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
		a *= 2
	}
	a--
}
`,
			wantLines: []int{7},
		},
		{
			name: "empty line after multiline if",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
		a *= 2
	}

	a--
}
`,
		},
		{
			name: "single line if",
			code: `package test
func run(a int) {
	if a > 0 { a++ }
	a--
}
`,
		},
		{
			name: "multiline if with else",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
	} else {
		a--
	}
	a *= 2
}
`,
			wantLines: []int{8},
		},
		{
			name: "multiline for range switch type switch and select",
			code: `package test
func run(values []int, v any, ch chan int) {
	for i := 0; i < 3; i++ {
		println(i)
	}
	for _, value := range values {
		println(value)
	}
	switch values[0] {
	case 1:
		println(1)
	}
	switch v.(type) {
	case int:
		println(2)
	}
	select {
	case <-ch:
		println(3)
	}
	println(4)
}
`,
			wantLines: []int{6, 9, 13, 17, 21},
		},
		{
			name: "last statement in block",
			code: `package test
func run(a int) {
	if a > 0 {
		if a > 1 {
			a++
		}
	}
}
`,
		},
		{
			name: "comment line is not an empty line",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
		a *= 2
	}
	// decrement
	a--
}
`,
			wantLines: []int{8},
		},
		{
			name: "empty line before comment",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
		a *= 2
	}

	// decrement
	a--
}
`,
		},
		{
			name: "statement in case body",
			code: `package test
func run(a int) {
	switch a {
	case 1:
		if a > 0 {
			a++
			a *= 2
		}
		a--
	}
}
`,
			wantLines: []int{9},
		},
	}

	runStatementSeparationTests(t, tests)
}

func TestStatementSeparationRule_GuardClauses(t *testing.T) {
	tests := []statementSeparationTest{
		{
			name: "guard clause before if",
			code: `package test
func run(a, b int) int {
	if a > 0 {
		a++
		return a
	}
	if b > 0 {
		b++
	}

	return b
}
`,
		},
		{
			name: "guard clause with continue, break, and panic",
			code: `package test
func run(values []int) {
	for _, value := range values {
		if value == 1 {
			println(value)
			continue
		}
		if value == 2 {
			println(value)
			break
		}
		if value == 3 {
			println(value)
			panic("three")
		}
		if value == 4 {
			println(value)
			continue
		}
		if value == 5 {
			println(value)
		}

		println(value)
	}
}
`,
		},
		{
			name: "guard clause before return",
			code: `package test
func run(a int) int {
	if a > 0 {
		a++
		return a
	}
	return 0
}
`,
			wantLines: []int{7},
		},
		{
			name: "guard clause before call",
			code: `package test
func run(a int) {
	if a > 0 {
		a++
		return
	}
	println(a)
}
`,
			wantLines: []int{7},
		},
		{
			name: "if with else before if",
			code: `package test
func run(a, b int) int {
	if a > 0 {
		return a
	} else {
		a++
	}
	if b > 0 {
		b++
	}

	return b
}
`,
			wantLines: []int{8},
		},
		{
			name: "if without terminating body before if",
			code: `package test
func run(a, b int) int {
	if a > 0 {
		a++
		a *= 2
	}
	if b > 0 {
		b++
	}

	return b
}
`,
			wantLines: []int{7},
		},
		{
			name: "fallthrough is not a guard clause",
			code: `package test
func run(a int) {
	switch a {
	case 1:
		if a > 0 {
			a++
			fallthrough
		}
		if a > 1 {
			a++
		}
	}
}
`,
			wantLines: []int{9},
		},
	}

	runStatementSeparationTests(t, tests)
}

func TestStatementSeparationRule_Branches(t *testing.T) {
	tests := []statementSeparationTest{
		{
			name: "return after three statements",
			code: `package test
func run(a, b, c int) int {
	a++
	b++
	c++
	return a + b + c
}
`,
			wantLines: []int{6},
		},
		{
			name: "return after two statements",
			code: `package test
func run(a, b int) int {
	a++
	b++
	return a + b
}
`,
		},
		{
			name: "return after one statement",
			code: `package test
func run(a int) int {
	a++
	return a
}
`,
		},
		{
			name: "empty line before return",
			code: `package test
func run(a, b, c int) int {
	a++
	b++
	c++

	return a + b + c
}
`,
		},
		{
			name: "break continue and goto",
			code: `package test
func run(values []int) {
	for _, value := range values {
		println(value)
		println(value)
		println(value)
		continue
	}

	for _, value := range values {
		println(value)
		println(value)
		println(value)
		break
	}

	for _, value := range values {
		println(value)
		println(value)
		println(value)
		goto done
	}

done:
}
`,
			wantLines: []int{7, 14, 21},
		},
		{
			name: "break and continue after two statements",
			code: `package test
func run(values []int) {
	for _, value := range values {
		println(value)
		println(value)
		continue
	}

	for _, value := range values {
		println(value)
		println(value)
		break
	}
}
`,
		},
		{
			name: "fallthrough",
			code: `package test
func run(a int) {
	switch a {
	case 1:
		println(a)
		println(a)
		println(a)
		fallthrough
	case 2:
		println(a)
	}
}
`,
		},
		{
			name: "return in case body",
			code: `package test
func run(a int) int {
	switch a {
	case 1:
		println(a)
		println(a)
		println(a)
		return a
	}

	return 0
}
`,
			wantLines: []int{8},
		},
		{
			name: "return in case body after two statements",
			code: `package test
func run(a int) int {
	switch a {
	case 1:
		println(a)
		println(a)
		return a
	}

	return 0
}
`,
		},
		{
			name: "return after multiline if reports once",
			code: `package test
func run(a, b int) int {
	b++
	if a > 0 {
		a++
		a *= 2
	}
	return a + b
}
`,
			wantLines: []int{8},
		},
		{
			name: "comment line is not an empty line",
			code: `package test
func run(a, b, c int) int {
	a++
	b++
	c++
	// sum
	return a + b + c
}
`,
			wantLines: []int{7},
		},
		{
			name: "function literal block",
			code: `package test
func run() func(a, b, c int) int {
	return func(a, b, c int) int {
		a++
		b++
		c++
		return a + b + c
	}
}
`,
			wantLines: []int{7},
		},
	}

	runStatementSeparationTests(t, tests)
}

func TestStatementSeparationRule_Metadata(t *testing.T) {
	rule := NewStatementSeparationRule()

	assert.Equal(t, "codequality-statement-separation", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
	assert.NotEmpty(t, rule.Description())
}

func runStatementSeparationTests(t *testing.T, tests []statementSeparationTest) {
	t.Helper()
	rule := NewStatementSeparationRule()

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
