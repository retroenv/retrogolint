package codequality

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestFuncOrderRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		code           string
		wantViolations int
	}{
		{
			name:     "full declaration order",
			filename: "test.go",
			code: `package test
type Service struct{}
func New() *Service { return nil }
func (s *Service) Run() {}
func (s *Service) validate() {}
func ExportedHelper() {}
type worker struct{}
func newWorker() *worker { return nil }
func (w *worker) run() {}
func helper() {}
`,
			wantViolations: 0,
		},
		{
			name:     "unexported interface directly before exported type is allowed when used",
			filename: "test.go",
			code: `package test
type serviceDep interface {
	Run()
}
type Service struct{
	dep serviceDep
}
func New() *Service { return nil }
`,
			wantViolations: 0,
		},
		{
			name:     "unexported non-interface type directly before exported type is allowed when used",
			filename: "test.go",
			code: `package test
type serviceDep func() bool
type Service struct{
	dep serviceDep
}
func New() *Service { return nil }
`,
			wantViolations: 0,
		},
		{
			name:     "unexported interface directly before exported type is not allowed when unused",
			filename: "test.go",
			code: `package test
type serviceDep interface {
	Run()
}
type Service struct{}
func New() *Service { return nil }
`,
			wantViolations: 1,
		},
		{
			name:     "unexported interface used by exported type cannot be below it",
			filename: "test.go",
			code: `package test
type Service struct{
	dep serviceDep
}
type serviceDep interface {
	Run()
}
func New() *Service { return nil }
`,
			wantViolations: 2,
		},
		{
			name:     "unexported non-interface type used by exported type cannot be below it",
			filename: "test.go",
			code: `package test
type Service struct{
	dep serviceDep
}
type serviceDep func() bool
func New() *Service { return nil }
`,
			wantViolations: 2,
		},
		{
			name:     "type after function",
			filename: "test.go",
			code: `package test
func New() *Service { return nil }
type Service struct{}
`,
			wantViolations: 1,
		},
		{
			name:     "exported method after unexported function",
			filename: "test.go",
			code: `package test
type Service struct{}
func helper() {}
func (s *Service) Run() {}
`,
			wantViolations: 1,
		},
		{
			name:     "unexported method after unexported function",
			filename: "test.go",
			code: `package test
type Service struct{}
func helper() {}
func (s *Service) validate() {}
`,
			wantViolations: 1,
		},
		{
			name:     "constructor after exported method",
			filename: "test.go",
			code: `package test
type Manager struct{}
type Bank struct{}
func (b *Bank) Get() {}
func New() *Manager { return nil }
func (m *Manager) Add() {}
`,
			wantViolations: 1,
		},
		{
			name:     "new prefix returning predeclared type is exported function",
			filename: "test.go",
			code: `package test
type Service struct{}
func NewCount() int { return 0 }
func (s *Service) Run() {}
`,
			wantViolations: 1,
		},
		{
			name:     "exported function after unexported type",
			filename: "test.go",
			code: `package test
type Service struct{}
type worker struct{}
func Build() {}
`,
			wantViolations: 1,
		},
		{
			name:     "method on unexported receiver is unexported",
			filename: "test.go",
			code: `package test
type Public struct{}
type service struct{}
func (s *service) Run() {}
func (p *Public) Run() {}
`,
			wantViolations: 1,
		},
		{
			name:     "test file skipped",
			filename: "service_test.go",
			code: `package test
func helper() {}
type mockService struct{}
func TestService() {}
`,
			wantViolations: 0,
		},
	}

	rule := NewFuncOrderRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tt.filename, tt.code, 0)
			assert.NoError(t, err)

			violations := rule.Check(fset, file)
			assert.Len(t, violations, tt.wantViolations)
		})
	}
}

func TestFuncOrderRule_Metadata(t *testing.T) {
	rule := NewFuncOrderRule()

	assert.Equal(t, "codequality-funcorder", rule.Name())
	assert.Equal(t, "codequality", rule.Category())
	assert.Equal(t, "Declarations should be ordered: exported types, exported constructors, exported methods, unexported methods on exported types, exported functions, unexported types, unexported constructors, methods on unexported types, unexported functions. Exception: unexported type dependencies may appear directly before an exported type when that type uses them.", rule.Description())
}
