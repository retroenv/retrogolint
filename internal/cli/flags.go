// Package cli provides command-line interface parsing and configuration.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/retroenv/retrogolib/cli"
	"github.com/retroenv/retrogolib/set"
	"github.com/retroenv/retrogolint/internal/linterconfig"
)

// ParseFlags parses command line flags and returns the configuration and paths to analyze.
func ParseFlags(version, commit, date string) (*linterconfig.Config, []string, error) {
	args := os.Args[1:]
	state, err := detectFlagOverrides(args)
	if err != nil {
		return nil, nil, err
	}

	cfg := linterconfig.DefaultConfig()
	if state.showHelp {
		flagSet, _, _, _ := buildFlagSet(cfg, state.configPath)
		flagSet.ShowUsage()
		os.Exit(0)
	}
	if state.showVersion {
		printVersion(version, commit, date)
		os.Exit(0)
	}

	fileCfg, err := linterconfig.LoadFileConfig(state.configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, nil, fmt.Errorf("failed to load config: %w", err)
		}
		if state.setFlags.Contains("config") {
			return nil, nil, fmt.Errorf("config file not found: %s", state.configPath)
		}
		// Config file doesn't exist and wasn't explicitly set - use defaults
	} else if err := linterconfig.ApplyFileConfig(cfg, fileCfg); err != nil {
		return nil, nil, fmt.Errorf("failed to apply config: %w", err)
	}

	flagSet, options, output, positional := buildFlagSet(cfg, state.configPath)

	if _, err := flagSet.Parse(args); err != nil {
		if errors.Is(err, cli.ErrHelpRequested) {
			flagSet.ShowUsage()
			os.Exit(0)
		}
		return nil, nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	if options.Version {
		printVersion(version, commit, date)
		os.Exit(0)
	}

	if err := applyFlagOverrides(cfg, state.setFlags, *options, *output); err != nil {
		return nil, nil, err
	}

	return cfg, positional.Packages, nil
}

type parameterFlags struct {
	Config string `flag:"config" usage:"Configuration file path"`
}

type optionFlags struct {
	Severity      string `flag:"severity" usage:"Minimum severity to report: error, warning, info"`
	Rules         string `flag:"rules" usage:"Comma-separated list of rule categories to check (default: all rules)"`
	DisabledRules string `flag:"disabled-rules" usage:"Comma-separated list of rules/categories to disable"`
	MaxPerRule    int    `flag:"max-per-rule" usage:"Maximum violations per rule (0 = unlimited)"`
	ExcludeTests  bool   `flag:"exclude-tests" usage:"Exclude test files from analysis"`
	ExcludeDirs   string `flag:"exclude-dirs" usage:"Comma-separated list of directories to exclude"`
	ExcludeFiles  string `flag:"exclude-files" usage:"Comma-separated list of file patterns to exclude"`
	Version       bool   `flag:"version" usage:"Show version information"`
}

type outputFlags struct {
	Format string `flag:"format" usage:"Output format: text, json"`
}

type positionalArgs struct {
	Packages []string `arg:"positional" usage:"packages to analyze"`
}

type flagState struct {
	setFlags    set.Set[string]
	configPath  string
	showHelp    bool
	showVersion bool
}

func printVersion(version, commit, date string) {
	versionString := version
	if commit != "" {
		if len(commit) > 7 {
			commit = commit[:7]
		}
		versionString += fmt.Sprintf(" (%s)", commit)
	}
	fmt.Printf("retrogolint version %s\n", versionString)
	if date != "" {
		fmt.Printf("build date: %s\n", date)
	}
}

func buildFlagSet(cfg *linterconfig.Config, configPath string) (*cli.FlagSet, *optionFlags, *outputFlags, *positionalArgs) {
	params := &parameterFlags{Config: configPath}
	options := &optionFlags{
		Severity:      cfg.Severity,
		Rules:         strings.Join(cfg.Rules, ","),
		DisabledRules: strings.Join(cfg.DisabledRules, ","),
		MaxPerRule:    cfg.MaxPerRule,
		ExcludeTests:  cfg.ExcludeTests,
		ExcludeDirs:   strings.Join(cfg.ExcludeDirs, ","),
		ExcludeFiles:  strings.Join(cfg.ExcludeFiles, ","),
	}
	output := &outputFlags{Format: cfg.Format}
	positional := &positionalArgs{}

	flagSet := cli.NewFlagSet("retrogolint")
	flagSet.AddSection("Parameters", params)
	flagSet.AddSection("Options", options)
	flagSet.AddSection("Output options", output)
	flagSet.AddPositional(positional)

	return flagSet, options, output, positional
}

func detectFlagOverrides(args []string) (flagState, error) {
	flags := flag.NewFlagSet("retrogolint", flag.ContinueOnError)
	flags.SetOutput(io.Discard)

	var configPath string
	var showVersion bool
	var showHelp bool
	var showShortHelp bool

	flags.StringVar(&configPath, "config", ".retrogolint.ini", "")
	flags.BoolVar(&showVersion, "version", false, "")
	flags.BoolVar(&showHelp, "help", false, "")
	flags.BoolVar(&showShortHelp, "h", false, "")

	for _, name := range []string{"format", "severity", "rules", "disabled-rules", "exclude-dirs", "exclude-files"} {
		flags.String(name, "", "")
	}
	flags.Int("max-per-rule", 0, "")
	flags.Bool("exclude-tests", false, "")

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return flagState{
				setFlags:   set.New[string](),
				configPath: configPath,
				showHelp:   true,
			}, nil
		}
		return flagState{}, fmt.Errorf("failed to parse flags: %w", err)
	}

	setFlags := set.New[string]()
	flags.Visit(func(flag *flag.Flag) {
		setFlags.Add(flag.Name)
	})

	return flagState{
		setFlags:    setFlags,
		configPath:  configPath,
		showHelp:    showHelp || showShortHelp,
		showVersion: showVersion,
	}, nil
}

func applyFlagOverrides(cfg *linterconfig.Config, setFlags set.Set[string], options optionFlags, output outputFlags) error {
	if setFlags.Contains("format") {
		if err := linterconfig.ValidateFormat(output.Format); err != nil {
			return fmt.Errorf("invalid format: %w", err)
		}
		cfg.Format = output.Format
	}
	if setFlags.Contains("severity") {
		switch options.Severity {
		case "info", "warning", "error":
			cfg.Severity = options.Severity
		default:
			return fmt.Errorf("invalid severity %q", options.Severity)
		}
	}
	if setFlags.Contains("rules") {
		cfg.Rules = linterconfig.ParseRules(options.Rules)
	}
	if setFlags.Contains("disabled-rules") {
		cfg.DisabledRules = linterconfig.ParseRules(options.DisabledRules)
	}
	if setFlags.Contains("max-per-rule") {
		if options.MaxPerRule < 0 {
			return fmt.Errorf("max-per-rule must be >= 0, got %d", options.MaxPerRule)
		}
		cfg.MaxPerRule = options.MaxPerRule
	}
	if setFlags.Contains("exclude-tests") {
		cfg.ExcludeTests = options.ExcludeTests
	}
	if setFlags.Contains("exclude-dirs") {
		cfg.ExcludeDirs = linterconfig.ParseRules(options.ExcludeDirs)
	}
	if setFlags.Contains("exclude-files") {
		cfg.ExcludeFiles = linterconfig.ParseRules(options.ExcludeFiles)
	}

	return nil
}
