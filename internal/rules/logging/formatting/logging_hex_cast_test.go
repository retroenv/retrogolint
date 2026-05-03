package loggingformatting

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestLoggingHexCastRule(t *testing.T) {
	tests := []struct {
		name           string
		code           string
		wantViolations int
	}{
		{
			name: "uint64 of uint16 should warn",
			code: `package test
func test(logger *Logger, address uint16) {
	logger.Debug("Handler", log.Hex("addr", uint64(uint16(address))))
}`,
			wantViolations: 1,
		},
		{
			name: "uint64 of uint32 should warn",
			code: `package test
func test(logger *Logger, value uint32) {
	logger.Debug("Handler", log.Hex("addr", uint64(uint32(value))))
}`,
			wantViolations: 1,
		},
		{
			name: "int conversion is allowed",
			code: `package test
func test(logger *Logger, address int16) {
	logger.Debug("Handler", log.Hex("addr", int(int16(address))))
}`,
			wantViolations: 0,
		},
		{
			name: "direct log.Hex value is fine",
			code: `package test
func test(logger *Logger, value uint16) {
	logger.Debug("Handler", log.Hex("addr", value))
}`,
			wantViolations: 0,
		},
		{
			name: "non-Hex field is ignored",
			code: `package test
func test(logger *Logger, value uint16) {
	logger.Debug("Handler", log.String("value", uint64(uint16(value))))
}`,
			wantViolations: 0,
		},
		{
			name: "single uint64 conversion without nested cast is ignored",
			code: `package test
func test(logger *Logger, value uint16) {
	logger.Debug("Handler", log.Hex("addr", uint64(value)))
}`,
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "test.go", tt.code, parser.ParseComments)
			assert.NoError(t, err)

			rule := NewLoggingHexCastRule()
			violations := rule.Check(fset, file)

			if len(violations) != tt.wantViolations {
				for i, v := range violations {
					t.Logf("Violation %d: %s at %v", i+1, v.Message, v.Position)
				}
			}
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}
