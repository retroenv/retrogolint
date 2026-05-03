package linterconfig

import (
	"path"
	"path/filepath"
	"strings"
)

const defaultRules = "" // Empty means all rules enabled

// RuleExclusions holds exclusion settings for a specific rule.
type RuleExclusions struct {
	ExcludeTests bool
	ExcludeDirs  []string
	ExcludeFiles []string
}

// RuleFileConfig represents per-rule exclusions from INI file.
type RuleFileConfig struct {
	ExcludeTests bool   `config:"exclude-tests"`
	ExcludeDirs  string `config:"exclude-dirs"`
	ExcludeFiles string `config:"exclude-files"`
}

// ShouldExcludeFile checks if a file should be excluded based on exclusion settings.
func (re *RuleExclusions) ShouldExcludeFile(filePath string) bool {
	if re.ExcludeTests && strings.HasSuffix(filePath, "_test.go") {
		return true
	}

	for _, dir := range re.ExcludeDirs {
		if strings.Contains(filePath, string(filepath.Separator)+dir+string(filepath.Separator)) ||
			strings.HasPrefix(filePath, dir+string(filepath.Separator)) {

			return true
		}
	}

	for _, pattern := range re.ExcludeFiles {
		if filePatternMatches(pattern, filePath) {
			return true
		}
	}

	return false
}

// ParseRules parses a comma-separated string of rules into a slice.
func ParseRules(s string) []string {
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	rules := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			rules = append(rules, trimmed)
		}
	}
	return rules
}

func filePatternMatches(pattern, filePath string) bool {
	pattern = filepath.ToSlash(pattern)
	filePath = filepath.ToSlash(filePath)

	if !strings.Contains(pattern, "/") {
		matched, err := path.Match(pattern, path.Base(filePath))
		return err == nil && matched
	}

	pattern = path.Clean(pattern)
	filePath = path.Clean(filePath)
	for {
		matched, err := path.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}

		trimmed := strings.TrimPrefix(filePath, "/")
		if trimmed == filePath {
			break
		}
		filePath = trimmed
	}

	for {
		idx := strings.IndexByte(filePath, '/')
		if idx == -1 {
			return false
		}
		filePath = filePath[idx+1:]

		matched, err := path.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}
	}
}

// extractRuleExclusions parses [rule.*] sections from INI data and returns per-rule exclusions.
func extractRuleExclusions(data []byte) map[string]RuleFileConfig {
	result := make(map[string]RuleFileConfig)
	lines := strings.Split(string(data), "\n")

	var currentRule string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if newRule, ok := parseRuleSectionHeader(trimmed); ok {
			currentRule = newRule
			result[currentRule] = RuleFileConfig{}
			continue
		}

		if shouldResetCurrentRule(trimmed) {
			currentRule = ""
			continue
		}

		if shouldSkipLine(trimmed, currentRule) {
			continue
		}

		key, value, ok := parseKeyValue(trimmed)
		if !ok {
			continue
		}

		updateRuleConfig(result, currentRule, key, value)
	}

	return result
}

// parseRuleSectionHeader extracts rule name from [rule.*] section headers.
func parseRuleSectionHeader(line string) (string, bool) {
	if !strings.HasPrefix(line, "[rule.") || !strings.HasSuffix(line, "]") {
		return "", false
	}
	ruleName := strings.TrimPrefix(line, "[rule.")
	ruleName = strings.TrimSuffix(ruleName, "]")
	if ruleName == "" {
		return "", false
	}
	return ruleName, true
}

// shouldResetCurrentRule checks if we should reset the current rule context.
func shouldResetCurrentRule(line string) bool {
	return strings.HasPrefix(line, "[") && !strings.HasPrefix(line, "[rule.")
}

// shouldSkipLine checks if a line should be skipped during parsing.
func shouldSkipLine(line, currentRule string) bool {
	if currentRule == "" {
		return true
	}
	if line == "" {
		return true
	}
	return strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";")
}

// parseKeyValue parses a key=value pair from a config line.
func parseKeyValue(line string) (key, value string, ok bool) {
	if !strings.Contains(line, "=") {
		return "", "", false
	}
	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])
	key = strings.ReplaceAll(key, "_", "-")
	return key, value, true
}

// updateRuleConfig updates the rule configuration with a key-value pair.
func updateRuleConfig(result map[string]RuleFileConfig, ruleName, key, value string) {
	ruleCfg := result[ruleName]
	switch key {
	case "exclude-tests":
		ruleCfg.ExcludeTests = value == "true" || value == "True" || value == "TRUE"
	case "exclude-dirs":
		ruleCfg.ExcludeDirs = value
	case "exclude-files":
		ruleCfg.ExcludeFiles = value
	}
	result[ruleName] = ruleCfg
}
