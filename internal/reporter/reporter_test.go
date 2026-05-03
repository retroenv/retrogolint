package reporter

import (
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/reporter/formats/jsonreport"
	"github.com/retroenv/retrogolint/internal/reporter/formats/textreport"
)

func TestNew(t *testing.T) {
	assert.IsType(t, &textreport.Reporter{}, New(""))
	assert.IsType(t, &textreport.Reporter{}, New("text"))
	assert.IsType(t, &jsonreport.Reporter{}, New("json"))
}
