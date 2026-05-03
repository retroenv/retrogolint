package testing

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestAssertUsageRule_Check(t *testing.T) {
	rule := NewAssertUsageRule()

	tests := []struct {
		name          string
		code          string
		wantViolation bool
		wantMessage   string
	}{
		{
			name: "NoError pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	err := foo()
	if err != nil {
		t.Fatal("error")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.NoError(t, err) instead of manual error check",
		},
		{
			name: "Error pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	err := foo()
	if err == nil {
		t.Fatal("expected error")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Error(t, err) instead of manual error check",
		},
		{
			name: "Equal pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	result := bar()
	if result != 42 {
		t.Error("wrong result")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Equal(t, 42, result) instead of manual equality check",
		},
		{
			name: "NotEqual pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	result := bar()
	if result == 0 {
		t.Error("result is zero")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.NotEqual(t, 0, result) instead of manual inequality check",
		},
		{
			name: "True pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	ok := check()
	if !ok {
		t.Error("not ok")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.True(t, ok) instead of manual boolean check",
		},
		{
			name: "False pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	bad := check()
	if bad {
		t.Error("is bad")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.False(t, bad) instead of manual boolean check",
		},
		{
			name: "NotNil pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	ptr := getPtr()
	if ptr == nil {
		t.Fatal("nil pointer")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.NotNil(t, ptr) instead of manual nil check",
		},
		{
			name: "Nil pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	ptr := getPtr()
	if ptr != nil {
		t.Error("expected nil")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Nil(t, ptr) instead of manual nil check",
		},
		{
			name: "Len pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	items := getItems()
	if len(items) != 5 {
		t.Error("wrong length")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Len(t, items, 5) instead of manual length check",
		},
		{
			name: "Greater pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	value := getValue()
	if value <= 10 {
		t.Error("too small")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Greater(t, value, 10) instead of manual comparison",
		},
		{
			name: "GreaterOrEqual pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	value := getValue()
	if value < 0 {
		t.Error("negative")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.GreaterOrEqual(t, value, 0) instead of manual comparison",
		},
		{
			name: "Less pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	value := getValue()
	if value >= 100 {
		t.Error("too large")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Less(t, value, 100) instead of manual comparison",
		},
		{
			name: "LessOrEqual pattern",
			code: `package test
import "testing"
func TestFoo(t *testing.T) {
	value := getValue()
	if value > 255 {
		t.Error("too large")
	}
}`,
			wantViolation: true,
			wantMessage:   "Use assert.LessOrEqual(t, value, 255) instead of manual comparison",
		},
		{
			name: "Non-test file",
			code: `package main
func main() {
	err := foo()
	if err != nil {
		panic(err)
	}
}`,
			wantViolation: false,
		},
		{
			name: "Non-test function",
			code: `package test
func helper() {
	err := foo()
	if err != nil {
		panic(err)
	}
}`,
			wantViolation: false,
		},
		{
			name: "assert.Equal with len() - expected first",
			code: `package test
import "testing"
import "github.com/retroenv/retrogolib/assert"
func TestFoo(t *testing.T) {
	output := []int{1, 2}
	assert.Equal(t, 2, len(output))
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Len(t, output, 2) instead of assert.Equal with len()",
		},
		{
			name: "assert.Equal with len() - len first",
			code: `package test
import "testing"
import "github.com/retroenv/retrogolib/assert"
func TestFoo(t *testing.T) {
	items := []string{"a", "b"}
	assert.Equal(t, len(items), 2)
}`,
			wantViolation: true,
			wantMessage:   "Use assert.Len(t, items, 2) instead of assert.Equal with len()",
		},
		{
			name: "assert.Equal without len() - should not trigger",
			code: `package test
import "testing"
import "github.com/retroenv/retrogolib/assert"
func TestFoo(t *testing.T) {
	value := 42
	assert.Equal(t, 42, value)
}`,
			wantViolation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, 0)
			assert.NoError(t, err)

			// Override filename for test file detection
			if tt.wantViolation {
				fset = token.NewFileSet()
				file, err = parser.ParseFile(fset, "test_test.go", tt.code, 0)
				assert.NoError(t, err)
			}

			violations := rule.Check(fset, file)

			if tt.wantViolation {
				assert.Len(t, violations, 1)
				assert.Equal(t, tt.wantMessage, violations[0].Message)
			} else {
				assert.Len(t, violations, 0)
			}
		})
	}
}

func TestAssertUsageRule_Metadata(t *testing.T) {
	rule := NewAssertUsageRule()

	assert.Equal(t, "assert-usage", rule.Name())
	assert.Equal(t, "Use assert package functions instead of manual test assertions", rule.Description())
	assert.Equal(t, "testing", rule.Category())
}
