// Package main is the entry point for the retrogolint linter tool.
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/retroenv/retrogolint/internal/analyzer"
	"github.com/retroenv/retrogolint/internal/cli"
	"github.com/retroenv/retrogolint/internal/reporter"
	"github.com/retroenv/retrogolint/internal/rules"
	"github.com/retroenv/retrogolint/internal/workflow"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

// errViolationsFound indicates violations were detected during analysis.
var errViolationsFound = errors.New("violations found")

func main() {
	if err := run(); err != nil {
		if errors.Is(err, errViolationsFound) {
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, paths, err := cli.ParseFlags(version, commit, date)
	if err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	registry := rules.NewRegistry()
	if err := registry.ValidateFilters(cfg.Rules); err != nil {
		return fmt.Errorf("invalid rules: %w", err)
	}
	if err := registry.ValidateFilters(cfg.DisabledRules); err != nil {
		return fmt.Errorf("invalid disabled rules: %w", err)
	}

	an := analyzer.New(cfg, registry)

	violations, err := workflow.Analyze(an, paths)
	if err != nil {
		return fmt.Errorf("analysis failed: %w", err)
	}

	rep := reporter.New(cfg.Format)
	if err := rep.Report(os.Stdout, violations); err != nil {
		return fmt.Errorf("failed to report violations: %w", err)
	}

	if len(violations) > 0 {
		return errViolationsFound
	}

	return nil
}
