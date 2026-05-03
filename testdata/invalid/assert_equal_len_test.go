package testdata

import (
	"testing"

	"github.com/retroenv/retrogolib/assert"
)

func TestAssertEqualLen(t *testing.T) {
	output := []int{1, 2}
	assert.Equal(t, 2, len(output))
}

func TestAssertEqualLenReversed(t *testing.T) {
	items := []string{"a", "b", "c"}
	assert.Equal(t, len(items), 3)
}

func TestAssertEqualLenWithMessage(t *testing.T) {
	data := []byte{1, 2, 3, 4}
	assert.Equal(t, 4, len(data), "data should have 4 bytes")
}

// This should not trigger - no len() call
func TestAssertEqualNoLen(t *testing.T) {
	value := 42
	assert.Equal(t, 42, value)
}
