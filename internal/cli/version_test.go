package cli

import (
	"os"
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolib/buildinfo"
)

func TestPrintVersion(t *testing.T) {
	output, err := os.CreateTemp(t.TempDir(), "version")
	assert.NoError(t, err)
	original := os.Stdout
	os.Stdout = output
	t.Cleanup(func() {
		os.Stdout = original
		assert.NoError(t, output.Close())
	})

	printVersion("1.0.14", "abcdef0123456789", "2026-11-17T02:03:53Z")
	data, err := os.ReadFile(output.Name())
	assert.NoError(t, err)
	want := "retrogolint version: v1.0.14\ncommit hash: abcdef0\ncommit time: 2026-11-17T02:03:53Z\n"
	if buildinfo.Read(buildinfo.Options{}).Modified {
		want += "source: modified\n"
	}

	assert.Equal(t, want, string(data))
}
