// Package workflow provides high-level orchestration of analysis operations.
package workflow

import (
	"fmt"

	"github.com/retroenv/retrogolint/internal/analyzer"
	"github.com/retroenv/retrogolint/internal/violation"
)

// Analyze performs analysis and returns violations.
func Analyze(
	a *analyzer.Analyzer,
	paths []string,
) ([]violation.Violation, error) {

	if len(paths) == 0 {
		paths = []string{"./..."}
	}

	violations, err := a.AnalyzePackages(paths)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze packages: %w", err)
	}

	return violations, nil
}
