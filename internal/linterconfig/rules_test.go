package linterconfig

import (
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestRuleExclusions_ShouldExcludeFile(t *testing.T) {
	tests := []struct {
		name       string
		exclusions RuleExclusions
		path       string
		want       bool
	}{
		{
			name:       "exclude test files",
			exclusions: RuleExclusions{ExcludeTests: true},
			path:       "test_file_test.go",
			want:       true,
		},
		{
			name:       "include test files",
			exclusions: RuleExclusions{ExcludeTests: false},
			path:       "test_file_test.go",
			want:       false,
		},
		{
			name:       "exclude directory",
			exclusions: RuleExclusions{ExcludeDirs: []string{"vendor"}},
			path:       "vendor/package/file.go",
			want:       true,
		},
		{
			name:       "include regular file",
			exclusions: RuleExclusions{ExcludeTests: false, ExcludeDirs: []string{}},
			path:       "pkg/file.go",
			want:       false,
		},
		{
			name:       "exclude file by basename glob",
			exclusions: RuleExclusions{ExcludeFiles: []string{"*_gen.go"}},
			path:       "pkg/generated/foo_gen.go",
			want:       true,
		},
		{
			name:       "exclude file by relative path",
			exclusions: RuleExclusions{ExcludeFiles: []string{"assert/assert_test.go"}},
			path:       "assert/assert_test.go",
			want:       true,
		},
		{
			name:       "exclude absolute file by relative path",
			exclusions: RuleExclusions{ExcludeFiles: []string{"assert/assert_test.go"}},
			path:       "/home/user/go/src/github.com/retroenv/retrogolib/assert/assert_test.go",
			want:       true,
		},
		{
			name:       "exclude absolute file by relative path glob",
			exclusions: RuleExclusions{ExcludeFiles: []string{"assert/*_test.go"}},
			path:       "/home/user/go/src/github.com/retroenv/retrogolib/assert/assert_test.go",
			want:       true,
		},
		{
			name:       "no match for different subdir with path pattern",
			exclusions: RuleExclusions{ExcludeFiles: []string{"assert/assert_test.go"}},
			path:       "other/assert_test.go",
			want:       false,
		},
		{
			name:       "invalid pattern does not exclude file",
			exclusions: RuleExclusions{ExcludeFiles: []string{"["}},
			path:       "pkg/file.go",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.exclusions.ShouldExcludeFile(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}
