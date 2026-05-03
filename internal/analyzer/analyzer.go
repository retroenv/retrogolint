// Package analyzer performs static analysis on Go source files.
package analyzer

import (
	"cmp"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"

	"github.com/retroenv/retrogolint/internal/linterconfig"
	"github.com/retroenv/retrogolint/internal/rules"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const goFileExt = ".go"

// Analyzer performs linting analysis on Go source files.
type Analyzer struct {
	config   *linterconfig.Config
	registry *rules.Registry
}

// New creates a new Analyzer with the given configuration and rule registry.
func New(cfg *linterconfig.Config, registry *rules.Registry) *Analyzer {
	return &Analyzer{
		config:   cfg,
		registry: registry,
	}
}

// FileSet returns the token.FileSet used by this analyzer.
func (a *Analyzer) FileSet() *token.FileSet {
	return token.NewFileSet()
}

// AnalyzePackages analyzes Go packages at the given paths.
// Paths can be files, directories, or Go package patterns like "./...".
func (a *Analyzer) AnalyzePackages(paths []string) ([]violation.Violation, error) {
	var allFiles []string

	for _, path := range paths {
		files, err := a.expandPath(path)
		if err != nil {
			return nil, fmt.Errorf("failed to expand path %s: %w", path, err)
		}
		allFiles = append(allFiles, files...)
	}

	return a.AnalyzeFiles(allFiles)
}

// AnalyzeFiles analyzes the specified Go source files.
func (a *Analyzer) AnalyzeFiles(paths []string) ([]violation.Violation, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	activeRules := a.registry.ActiveRules(a.config.Rules, a.config.DisabledRules)
	if len(activeRules) == 0 {
		return nil, nil
	}

	allViolations, err := a.analyzeFilesParallel(paths, activeRules)
	if err != nil {
		return nil, err
	}

	// Apply per-rule exclusions before severity filtering
	filteredByExclusions := a.filterByRuleExclusions(allViolations)

	return a.filterViolations(filteredByExclusions), nil
}

// analyzeFile analyzes a single Go source file.
func (a *Analyzer) analyzeFile(path string, activeRules []api.Rule) ([]violation.Violation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var violations []violation.Violation

	for _, rule := range activeRules {
		ruleViolations := rule.Check(fset, file)
		violations = append(violations, ruleViolations...)
	}

	return violations, nil
}

// analyzeFilesParallel analyzes files in parallel using a worker pool.
func (a *Analyzer) analyzeFilesParallel(paths []string, activeRules []api.Rule) ([]violation.Violation, error) {
	var allViolations []violation.Violation
	var allErrors []error

	workerCount := runtime.GOMAXPROCS(0)
	if workerCount <= 0 {
		workerCount = 1
	}

	jobs := make(chan string)
	results := make(chan fileAnalysisResult, workerCount*2)

	var wg sync.WaitGroup
	for range workerCount {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				violations, err := a.analyzeFile(path, activeRules)
				if err != nil {
					results <- fileAnalysisResult{err: fmt.Errorf("failed to analyze %s: %w", path, err)}
					continue
				}
				if len(violations) > 0 {
					results <- fileAnalysisResult{violations: violations}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var collector sync.WaitGroup
	collector.Add(1)
	go func() {
		defer collector.Done()
		for result := range results {
			if result.err != nil {
				allErrors = append(allErrors, result.err)
				continue
			}
			allViolations = append(allViolations, result.violations...)
		}
	}()

	for _, path := range paths {
		if a.config.ShouldSkipPath(path) {
			continue
		}
		jobs <- path
	}
	close(jobs)

	collector.Wait()
	return allViolations, errors.Join(allErrors...)
}

// filterViolations filters violations by severity and applies per-rule limits.
func (a *Analyzer) filterViolations(violations []violation.Violation) []violation.Violation {
	minSeverity := a.config.MinSeverity()
	filtered := make([]violation.Violation, 0, len(violations))
	for _, v := range violations {
		if v.ShouldReport(minSeverity) {
			filtered = append(filtered, v)
		}
	}

	sortViolations(filtered)

	if a.config.MaxPerRule > 0 {
		filtered = a.limitViolationsPerRule(filtered)
	}

	return filtered
}

// expandPath expands a path pattern into a list of Go source files.
func (a *Analyzer) expandPath(path string) ([]string, error) {
	if path == "./..." {
		return a.findGoFiles(".")
	}

	if len(path) > 4 && path[len(path)-4:] == "/..." {
		dir := path[:len(path)-4]
		if dir == "" {
			dir = "."
		}
		return a.findGoFiles(dir)
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path %s: %w", path, err)
	}

	if info.IsDir() {
		return a.findGoFilesInDir(path)
	}

	if filepath.Ext(path) == goFileExt {
		return []string{path}, nil
	}

	return nil, fmt.Errorf("not a Go file: %s", path)
}

// findGoFiles recursively finds all .go files in a directory and its subdirectories.
func (a *Analyzer) findGoFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip directories starting with . (except "." itself).
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}

			if slices.Contains(a.config.ExcludeDirs, info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if filepath.Ext(path) == goFileExt && !a.config.ShouldExcludeFile(path) {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return files, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}
	return files, nil
}

// findGoFilesInDir finds all .go files in a directory (non-recursive).
func (a *Analyzer) findGoFilesInDir(dir string) ([]string, error) {
	var files []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		if filepath.Ext(path) == goFileExt && !a.config.ShouldExcludeFile(path) {
			files = append(files, path)
		}
	}

	return files, nil
}

// limitViolationsPerRule limits the number of violations reported per rule.
func (a *Analyzer) limitViolationsPerRule(violations []violation.Violation) []violation.Violation {
	counts := make(map[string]int)
	result := make([]violation.Violation, 0, len(violations))

	for _, v := range violations {
		if counts[v.Rule] < a.config.MaxPerRule {
			result = append(result, v)
			counts[v.Rule]++
		}
	}

	return result
}

// filterByRuleExclusions filters violations based on per-rule exclusion settings.
func (a *Analyzer) filterByRuleExclusions(violations []violation.Violation) []violation.Violation {
	if len(a.config.PerRuleExclusions) == 0 {
		return violations // No per-rule exclusions configured
	}

	filtered := make([]violation.Violation, 0, len(violations))
	for _, v := range violations {
		if !a.shouldExcludeViolation(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// shouldExcludeViolation checks if a violation should be excluded based on per-rule settings.
func (a *Analyzer) shouldExcludeViolation(v violation.Violation) bool {
	if ruleExcl, ok := a.config.PerRuleExclusions[v.Rule]; ok {
		if ruleExcl.ShouldExcludeFile(v.Position.Filename) {
			return true
		}
	}

	if category := a.getRuleCategory(v.Rule); category != "" {
		if catExcl, ok := a.config.PerRuleExclusions[category]; ok {
			if catExcl.ShouldExcludeFile(v.Position.Filename) {
				return true
			}
		}
	}

	return false
}

// getRuleCategory looks up the category for a given rule name.
func (a *Analyzer) getRuleCategory(ruleName string) string {
	rule, ok := a.registry.Get(ruleName)
	if !ok {
		return ""
	}
	return rule.Category()
}

type fileAnalysisResult struct {
	violations []violation.Violation
	err        error
}

func sortViolations(violations []violation.Violation) {
	slices.SortFunc(violations, func(a, b violation.Violation) int {
		if c := cmp.Compare(a.Position.Filename, b.Position.Filename); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Position.Line, b.Position.Line); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Position.Column, b.Position.Column); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Rule, b.Rule); c != 0 {
			return c
		}
		return cmp.Compare(a.Message, b.Message)
	})
}
