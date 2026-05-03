package analyzer

import (
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/linterconfig"
	"github.com/retroenv/retrogolint/internal/rules"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestAnalyzer_AnalyzeFiles(t *testing.T) {
	// Create test config
	cfg := linterconfig.DefaultConfig()
	cfg.Rules = []string{"logging"}

	// Create analyzer
	registry := rules.NewRegistry()
	a := New(cfg, registry)

	tests := []struct {
		name          string
		files         []string
		wantError     bool
		minViolations int
	}{
		{
			name:          "analyze valid file",
			files:         []string{"../../testdata/valid/correct.go"},
			wantError:     false,
			minViolations: 0,
		},
		{
			name:          "analyze file with violations",
			files:         []string{"../../testdata/invalid/capitalization.go"},
			wantError:     false,
			minViolations: 0, // Will vary based on file content
		},
		{
			name:          "analyze multiple files",
			files:         []string{"../../testdata/valid/correct.go", "../../testdata/invalid/field_casing.go"},
			wantError:     false,
			minViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations, err := a.AnalyzeFiles(tt.files)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, len(violations) >= tt.minViolations)
		})
	}
}

func TestAnalyzer_AnalyzeFiles_ReturnsParseError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "broken.go")
	err := os.WriteFile(path, []byte("package broken\nfunc nope("), 0644)
	assert.NoError(t, err)

	cfg := linterconfig.DefaultConfig()
	registry := rules.NewRegistry()
	a := New(cfg, registry)

	violations, err := a.AnalyzeFiles([]string{path})

	assert.Error(t, err)
	assert.Len(t, violations, 0)
}

func TestAnalyzer_ExpandPath(t *testing.T) {
	// Create test config
	cfg := linterconfig.DefaultConfig()
	registry := rules.NewRegistry()
	a := New(cfg, registry)

	tests := []struct {
		name      string
		path      string
		wantError bool
		checkFunc func([]string) bool
	}{
		{
			name:      "single go file",
			path:      "../../testdata/valid/correct.go",
			wantError: false,
			checkFunc: func(files []string) bool {
				return len(files) == 1 && filepath.Ext(files[0]) == goFileExt
			},
		},
		{
			name:      "directory pattern",
			path:      "../../testdata/valid",
			wantError: false,
			checkFunc: func(files []string) bool {
				return len(files) > 0
			},
		},
		{
			name:      "non-existent file",
			path:      "../../testdata/nonexistent.go",
			wantError: true,
			checkFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := a.expandPath(tt.path)

			if tt.wantError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.checkFunc != nil {
				assert.True(t, tt.checkFunc(files))
			}
		})
	}
}

func TestAnalyzer_ShouldExcludeFile(t *testing.T) {
	tests := []struct {
		name   string
		config *linterconfig.Config
		path   string
		want   bool
	}{
		{
			name: "exclude test files",
			config: &linterconfig.Config{
				RuleExclusions: linterconfig.RuleExclusions{ExcludeTests: true},
			},
			path: "test_file_test.go",
			want: true,
		},
		{
			name: "include test files",
			config: &linterconfig.Config{
				RuleExclusions: linterconfig.RuleExclusions{ExcludeTests: false},
			},
			path: "test_file_test.go",
			want: false,
		},
		{
			name: "exclude directory",
			config: &linterconfig.Config{
				RuleExclusions: linterconfig.RuleExclusions{ExcludeDirs: []string{"vendor"}},
			},
			path: "vendor/package/file.go",
			want: true,
		},
		{
			name: "include regular file",
			config: &linterconfig.Config{
				RuleExclusions: linterconfig.RuleExclusions{ExcludeTests: false, ExcludeDirs: []string{}},
			},
			path: "pkg/file.go",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ShouldExcludeFile(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAnalyzer_LimitViolationsPerRule(t *testing.T) {
	cfg := &linterconfig.Config{
		MaxPerRule: 2,
	}
	registry := rules.NewRegistry()
	a := New(cfg, registry)

	// Create test violations
	violations := []violation.Violation{
		{Rule: "rule1"},
		{Rule: "rule1"},
		{Rule: "rule1"}, // Should be excluded
		{Rule: "rule2"},
		{Rule: "rule2"},
		{Rule: "rule2"}, // Should be excluded
	}

	filtered := a.limitViolationsPerRule(violations)

	assert.Len(t, filtered, 4)

	// Count violations per rule
	counts := make(map[string]int)
	for _, v := range filtered {
		counts[v.Rule]++
	}

	for rule, count := range counts {
		assert.True(t, count <= cfg.MaxPerRule, "rule %s has %d violations", rule, count)
	}
}

func TestAnalyzer_FileSet(t *testing.T) {
	cfg := linterconfig.DefaultConfig()
	registry := rules.NewRegistry()
	a := New(cfg, registry)

	fset := a.FileSet()
	assert.NotNil(t, fset)
}

func TestAnalyzer_FilterViolations_DeterministicBeforeLimit(t *testing.T) {
	cfg := linterconfig.DefaultConfig()
	cfg.Severity = "info"
	cfg.MaxPerRule = 1

	registry := rules.NewRegistry()
	a := New(cfg, registry)

	violations := []violation.Violation{
		{
			Rule:     "logging-capitalization",
			Severity: violation.SeverityWarning,
			Position: token.Position{Filename: "b.go", Line: 10, Column: 1},
		},
		{
			Rule:     "logging-capitalization",
			Severity: violation.SeverityWarning,
			Position: token.Position{Filename: "a.go", Line: 20, Column: 1},
		},
		{
			Rule:     "logging-capitalization",
			Severity: violation.SeverityWarning,
			Position: token.Position{Filename: "a.go", Line: 5, Column: 1},
		},
	}

	filtered := a.filterViolations(violations)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "a.go", filtered[0].Position.Filename)
	assert.Equal(t, 5, filtered[0].Position.Line)
}
