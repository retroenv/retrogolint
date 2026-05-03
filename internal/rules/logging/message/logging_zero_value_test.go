package loggingmessage

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestLoggingZeroValueRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		code      string
		wantCount int
	}{
		{
			name: "detects &log.Logger{}",
			code: `package main

import "log"

func main() {
	logger := &log.Logger{}
	logger.Info("test")
}`,
			wantCount: 1,
		},
		{
			name: "detects multiple instances",
			code: `package main

import "log"

func main() {
	logger1 := &log.Logger{}
	logger2 := &log.Logger{}
	logger1.Info("test1")
	logger2.Info("test2")
}`,
			wantCount: 2,
		},
		{
			name: "ignores nil logger",
			code: `package main

import "log"

func main() {
	var logger *log.Logger
	logger.Info("test")
}`,
			wantCount: 0,
		},
		{
			name: "ignores non-empty composite literal",
			code: `package main

import "log"

func main() {
	logger := &log.Logger{Writer: os.Stdout}
	logger.Info("test")
}`,
			wantCount: 0,
		},
		{
			name: "ignores other types",
			code: `package main

type MyStruct struct{}

func main() {
	s := &MyStruct{}
	_ = s
}`,
			wantCount: 0,
		},
		{
			name: "ignores non-log.Logger types",
			code: `package main

import "mylog"

func main() {
	logger := &mylog.Logger{}
	logger.Info("test")
}`,
			wantCount: 0,
		},
		{
			name: "detects in variable declaration",
			code: `package main

import "log"

var logger = &log.Logger{}
`,
			wantCount: 1,
		},
		{
			name: "detects in function return",
			code: `package main

import "log"

func getLogger() *log.Logger {
	return &log.Logger{}
}`,
			wantCount: 1,
		},
		{
			name: "detects in struct field assignment",
			code: `package main

import "log"

type App struct {
	Logger *log.Logger
}

func main() {
	app := App{
		Logger: &log.Logger{},
	}
	_ = app
}`,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingZeroValueRule()
			violations := rule.Check(fset, file)

			assert.Len(t, violations, tt.wantCount)
		})
	}
}
