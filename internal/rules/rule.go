// Package rules wires all lint rules and helpers into a single registry.
package rules

import (
	"fmt"
	"slices"
	"strings"

	"github.com/retroenv/retrogolib/set"
	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/rules/codequality"
	"github.com/retroenv/retrogolint/internal/rules/collections"
	loggingfields "github.com/retroenv/retrogolint/internal/rules/logging/fields"
	loggingformatting "github.com/retroenv/retrogolint/internal/rules/logging/formatting"
	loggingmessage "github.com/retroenv/retrogolint/internal/rules/logging/message"
	loggingperformance "github.com/retroenv/retrogolint/internal/rules/logging/performance"
	linttesting "github.com/retroenv/retrogolint/internal/rules/testing"
)

// Registry holds all registered rules.
type Registry struct {
	rules map[string]api.Rule
}

// NewRegistry creates a new rule registry with all built-in rules registered.
func NewRegistry() *Registry {
	r := &Registry{rules: make(map[string]api.Rule)}

	r.Register(loggingmessage.NewLoggingCapitalizationRule())
	r.Register(loggingfields.NewLoggingFieldCasingRule())
	r.Register(loggingmessage.NewLoggingNilCheckRule())
	r.Register(loggingmessage.NewLoggingZeroValueRule())
	r.Register(loggingperformance.NewLoggingMethodCallsRule())
	r.Register(loggingformatting.NewLoggingHexFormattingRule())
	r.Register(loggingformatting.NewLoggingHexCastRule())
	r.Register(loggingformatting.NewLoggingTypeFormattingRule())
	r.Register(loggingformatting.NewLoggingFmtInFieldRule())
	r.Register(loggingperformance.NewLoggingEfficiencyRule())
	r.Register(loggingfields.NewLoggingSpecializedFieldRule())
	r.Register(collections.NewMapSetRule())
	r.Register(collections.NewSetMembershipRule())
	r.Register(collections.NewSetProjectionRule())
	r.Register(collections.NewSetSortRule())
	r.Register(linttesting.NewAssertUsageRule())
	r.Register(codequality.NewFuncOrderRule())
	r.Register(codequality.NewParamPriorityRule())
	r.Register(codequality.NewTypeStutterRule())

	return r
}

// Register adds a rule to the registry.
func (r *Registry) Register(rule api.Rule) {
	r.rules[rule.Name()] = rule
}

// Get retrieves a rule by name.
// nolint: ireturn
func (r *Registry) Get(name string) (api.Rule, bool) {
	rule, ok := r.rules[name]
	return rule, ok
}

// All returns all registered rules.
func (r *Registry) All() []api.Rule {
	rules := make([]api.Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// ActiveRules returns the rules that should be checked based on the provided rule filters.
func (r *Registry) ActiveRules(ruleFilters, disabled []string) []api.Rule {
	disabledRules := set.NewFromSlice(disabled)

	if len(ruleFilters) == 0 {
		allRules := r.All()
		activeRules := make([]api.Rule, 0, len(allRules))
		for _, rule := range allRules {
			if isRuleDisabled(rule, disabledRules) {
				continue
			}
			activeRules = append(activeRules, rule)
		}
		return activeRules
	}

	activeRules := make([]api.Rule, 0)
	ruleMap := set.New[string]()

	for _, filter := range ruleFilters {
		if rule, ok := r.Get(filter); ok {
			if !ruleMap.Contains(filter) && !isRuleDisabled(rule, disabledRules) {
				activeRules = append(activeRules, rule)
				ruleMap.Add(filter)
			}
			continue
		}

		for _, rule := range r.All() {
			if rule.Category() == filter && !ruleMap.Contains(rule.Name()) && !isRuleDisabled(rule, disabledRules) {
				activeRules = append(activeRules, rule)
				ruleMap.Add(rule.Name())
			}
		}
	}

	return activeRules
}

// ValidateFilters returns an error when a rule filter is not a known rule or category.
func (r *Registry) ValidateFilters(filters []string) error {
	if len(filters) == 0 {
		return nil
	}

	validCategories := set.New[string]()
	for _, rule := range r.All() {
		validCategories.Add(rule.Category())
	}

	unknown := make([]string, 0)
	for _, filter := range filters {
		if _, ok := r.Get(filter); ok {
			continue
		}
		if validCategories.Contains(filter) {
			continue
		}
		unknown = append(unknown, filter)
	}

	if len(unknown) == 0 {
		return nil
	}

	slices.Sort(unknown)
	return fmt.Errorf("unknown rule or category: %s", strings.Join(unknown, ", "))
}

func isRuleDisabled(rule api.Rule, disabled set.Set[string]) bool {
	return disabled.Contains(rule.Name()) || disabled.Contains(rule.Category())
}
