// Package linterconfig provides configuration management for the linter.
package linterconfig

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/retroenv/retrogolib/config"
	"github.com/retroenv/retrogolint/internal/violation"
)

const (
	defaultFormat   = "text"
	defaultSeverity = "warning"
)

// Config holds the linter configuration.
type Config struct {
	// Output configuration
	Format   string // Output format: text, json
	Severity string // Minimum severity to report: error, warning, info

	// Rule configuration
	Rules         []string // Comma-separated rule categories or specific rules
	DisabledRules []string // Rules to disable

	// File filtering (embedded for shared ShouldExcludeFile logic)
	RuleExclusions

	// Per-rule exclusions (key is rule name or category)
	PerRuleExclusions map[string]*RuleExclusions

	// MaxPerRule limits the number of violations reported per rule (0 = unlimited)
	MaxPerRule int
}

// FileConfig describes configuration values loaded from a file.
type FileConfig struct {
	Format        string `config:"format"`
	Severity      string `config:"severity"`
	Rules         string `config:"rules"`
	DisabledRules string `config:"disabled-rules"`
	MaxPerRule    int    `config:"max-per-rule"`
	ExcludeTests  bool   `config:"exclude-tests"`
	ExcludeDirs   string `config:"exclude-dirs"`
	ExcludeFiles  string `config:"exclude-files"`

	// Per-rule exclusions (key is rule name or category)
	RuleExclusions map[string]RuleFileConfig
}

// MinSeverity returns the minimum severity as a violation.Severity value.
func (c *Config) MinSeverity() violation.Severity {
	return violation.ParseSeverity(c.Severity)
}

// ShouldSkipPath checks if a path should be skipped (doesn't exist or is excluded).
func (c *Config) ShouldSkipPath(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return true
	}

	if info.IsDir() {
		return false
	}

	if !strings.HasSuffix(path, ".go") {
		return true
	}

	return c.ShouldExcludeFile(path)
}

// ValidateFormat checks whether an output format is supported.
func ValidateFormat(format string) error {
	switch format {
	case "text", "json":
		return nil
	case "":
		return errors.New("format cannot be empty")
	default:
		return fmt.Errorf("invalid format %q", format)
	}
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Format:            defaultFormat,
		Severity:          defaultSeverity,
		Rules:             []string{}, // Empty means all rules enabled
		MaxPerRule:        0,
		PerRuleExclusions: make(map[string]*RuleExclusions),
	}
}

// DefaultFileConfig returns default values for file-based configuration.
func DefaultFileConfig() FileConfig {
	return FileConfig{
		Format:         defaultFormat,
		Severity:       defaultSeverity,
		Rules:          defaultRules,
		MaxPerRule:     0,
		RuleExclusions: make(map[string]RuleFileConfig),
	}
}

// LoadFileConfig loads configuration values from a file.
func LoadFileConfig(path string) (FileConfig, error) {
	if path == "" {
		return DefaultFileConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return FileConfig{}, fmt.Errorf("failed to read config file: %w", err)
	}

	fileCfg := DefaultFileConfig()
	normalized := normalizeConfigData(data)
	parsed, err := config.LoadConfigBytes(normalized)
	if err != nil {
		return FileConfig{}, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}
	if err := parsed.Unmarshal(&fileCfg); err != nil {
		return FileConfig{}, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	fileCfg.RuleExclusions = extractRuleExclusions(data)

	return fileCfg, nil
}

// ApplyFileConfig applies file-based configuration values to the runtime config.
func ApplyFileConfig(cfg *Config, fileCfg FileConfig) error {
	if err := ValidateFormat(fileCfg.Format); err != nil {
		return err
	}
	cfg.Format = fileCfg.Format

	switch fileCfg.Severity {
	case "info", "warning", "error":
		cfg.Severity = fileCfg.Severity
	default:
		return fmt.Errorf("invalid severity %q", fileCfg.Severity)
	}

	cfg.Rules = ParseRules(fileCfg.Rules)
	cfg.DisabledRules = ParseRules(fileCfg.DisabledRules)

	if fileCfg.MaxPerRule < 0 {
		return fmt.Errorf("max-per-rule must be >= 0, got %d", fileCfg.MaxPerRule)
	}
	cfg.MaxPerRule = fileCfg.MaxPerRule
	cfg.ExcludeTests = fileCfg.ExcludeTests
	cfg.ExcludeDirs = ParseRules(fileCfg.ExcludeDirs)
	cfg.ExcludeFiles = ParseRules(fileCfg.ExcludeFiles)

	if len(fileCfg.RuleExclusions) > 0 {
		cfg.PerRuleExclusions = make(map[string]*RuleExclusions)
		for ruleName, ruleFileCfg := range fileCfg.RuleExclusions {
			cfg.PerRuleExclusions[ruleName] = &RuleExclusions{
				ExcludeTests: ruleFileCfg.ExcludeTests,
				ExcludeDirs:  ParseRules(ruleFileCfg.ExcludeDirs),
				ExcludeFiles: ParseRules(ruleFileCfg.ExcludeFiles),
			}
		}
	}

	return nil
}

func normalizeConfigData(data []byte) []byte {
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		trimmedLeft := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmedLeft, ";") {
			prefixLen := len(line) - len(trimmedLeft)
			line = line[:prefixLen] + "#" + trimmedLeft[1:]
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "[") || !strings.Contains(trimmed, "=") {
			lines[i] = line
			continue
		}

		eq := strings.Index(line, "=")
		if eq == -1 {
			lines[i] = line
			continue
		}

		keyPart := line[:eq]
		keyPart = strings.ReplaceAll(keyPart, "_", "-")
		lines[i] = keyPart + line[eq:]
	}

	return []byte(strings.Join(lines, "\n"))
}
