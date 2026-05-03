package cli

import (
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolib/set"
	"github.com/retroenv/retrogolint/internal/linterconfig"
)

func TestDetectFlagOverrides_DisabledRulesFlag(t *testing.T) {
	state, err := detectFlagOverrides([]string{"-disabled-rules", "logging,collections-map-set", "./..."})
	assert.NoError(t, err)
	assert.True(t, state.setFlags.Contains("disabled-rules"))
}

func TestApplyFlagOverrides_DisabledRules(t *testing.T) {
	cfg := linterconfig.DefaultConfig()
	setFlags := set.New[string]()
	setFlags.Add("disabled-rules")

	err := applyFlagOverrides(cfg, setFlags, optionFlags{
		DisabledRules: "logging,collections-map-set",
	}, outputFlags{})

	assert.NoError(t, err)
	assert.Equal(t, []string{"logging", "collections-map-set"}, cfg.DisabledRules)
}

func TestApplyFlagOverrides_InvalidFormat(t *testing.T) {
	cfg := linterconfig.DefaultConfig()
	setFlags := set.New[string]()
	setFlags.Add("format")

	err := applyFlagOverrides(cfg, setFlags, optionFlags{}, outputFlags{
		Format: "yaml",
	})

	assert.Error(t, err)
}
