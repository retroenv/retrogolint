package linterconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestLoadFileConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "retrogolint.ini")

	content := `; comment
format = json
severity = info
rules = logging,logging-capitalization
disabled_rules = logging-efficiency
max_per_rule = 3
exclude_tests = true
exclude_dirs = vendor,testdata
exclude_files = *_gen.go
`

	err := os.WriteFile(path, []byte(content), 0644)
	assert.NoError(t, err)

	fileCfg, err := LoadFileConfig(path)
	assert.NoError(t, err)
	assert.Equal(t, "json", fileCfg.Format)
	assert.Equal(t, "info", fileCfg.Severity)
	assert.Equal(t, "logging,logging-capitalization", fileCfg.Rules)
	assert.Equal(t, "logging-efficiency", fileCfg.DisabledRules)
	assert.Equal(t, 3, fileCfg.MaxPerRule)
	assert.True(t, fileCfg.ExcludeTests)
	assert.Equal(t, "vendor,testdata", fileCfg.ExcludeDirs)
	assert.Equal(t, "*_gen.go", fileCfg.ExcludeFiles)

	cfg := DefaultConfig()
	err = ApplyFileConfig(cfg, fileCfg)
	assert.NoError(t, err)
	assert.Equal(t, "json", cfg.Format)
	assert.Equal(t, "info", cfg.Severity)
	assert.Equal(t, []string{"logging", "logging-capitalization"}, cfg.Rules)
	assert.Equal(t, []string{"logging-efficiency"}, cfg.DisabledRules)
	assert.Equal(t, 3, cfg.MaxPerRule)
	assert.True(t, cfg.ExcludeTests)
	assert.Equal(t, []string{"vendor", "testdata"}, cfg.ExcludeDirs)
	assert.Equal(t, []string{"*_gen.go"}, cfg.ExcludeFiles)
}

func TestLoadFileConfig_WithGlobalSections(t *testing.T) {
	content := `# retrogolint configuration

[output]
format = text
severity = info
max_per_rule = 0

[rules]
enable_all = true
disabled-rules = logging-efficiency

[integration]
exclude_tests = true
exclude_dirs = vendor,testdata
exclude_files = *_gen.go
`

	fileCfg, err := loadFileConfigBytes([]byte(content), "testdata")
	assert.NoError(t, err)
	assert.Equal(t, "text", fileCfg.Format)
	assert.Equal(t, "info", fileCfg.Severity)
	assert.Equal(t, "logging-efficiency", fileCfg.DisabledRules)
	assert.Equal(t, 0, fileCfg.MaxPerRule)
	assert.True(t, fileCfg.ExcludeTests)
	assert.Equal(t, "vendor,testdata", fileCfg.ExcludeDirs)
	assert.Equal(t, "*_gen.go", fileCfg.ExcludeFiles)
}

func TestLoadFileConfig_PerRuleExclusions(t *testing.T) {
	content := `[rule.logging-capitalization]
exclude-files = assert/*_test.go
exclude-dirs = generated

[rule.logging]
exclude-tests = true
exclude_files = legacy_*.go
`

	fileCfg, err := loadFileConfigBytes([]byte(content), "testdata")
	assert.NoError(t, err)

	ruleFileCfg := fileCfg.RuleExclusions["logging-capitalization"]
	assert.Equal(t, "assert/*_test.go", ruleFileCfg.ExcludeFiles)
	assert.Equal(t, "generated", ruleFileCfg.ExcludeDirs)

	categoryFileCfg := fileCfg.RuleExclusions["logging"]
	assert.True(t, categoryFileCfg.ExcludeTests)
	assert.Equal(t, "legacy_*.go", categoryFileCfg.ExcludeFiles)

	cfg := DefaultConfig()
	err = ApplyFileConfig(cfg, fileCfg)
	assert.NoError(t, err)

	ruleExcl := cfg.PerRuleExclusions["logging-capitalization"]
	assert.NotNil(t, ruleExcl)
	assert.Equal(t, []string{"assert/*_test.go"}, ruleExcl.ExcludeFiles)
	assert.Equal(t, []string{"generated"}, ruleExcl.ExcludeDirs)

	categoryExcl := cfg.PerRuleExclusions["logging"]
	assert.NotNil(t, categoryExcl)
	assert.True(t, categoryExcl.ExcludeTests)
	assert.Equal(t, []string{"legacy_*.go"}, categoryExcl.ExcludeFiles)
}

func TestApplyFileConfig_InvalidFormat(t *testing.T) {
	cfg := DefaultConfig()
	fileCfg := DefaultFileConfig()
	fileCfg.Format = "yaml"

	err := ApplyFileConfig(cfg, fileCfg)
	assert.Error(t, err)
}
