package markdown

import (
	"testing"

	"github.com/retroenv/retrogolib/assert"
	"github.com/retroenv/retrogolint/internal/violation"
)

func TestTableStructureRule_Metadata(t *testing.T) {
	rule := NewTableStructureRule()

	assert.Equal(t, "markdown-table-structure", rule.Name())
	assert.Equal(t, "markdown", rule.Category())
	assert.Equal(t, violation.SeverityWarning, rule.Severity())
}

func TestTableStructureRule_CheckFile(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantViolations int
		wantLine       int
	}{
		{
			name: "consistent table",
			content: "# Title\n\n" +
				"| Stage | Dependency | Deliverable |\n" +
				"| --- | --- | --- |\n" +
				"| 00 | None | Verified baseline |\n",
			wantViolations: 0,
		},
		{
			name: "separator row with a missing cell",
			content: "| Stage | Dependency | Deliverable |\n" +
				"| --- | --- |\n" +
				"| 00 | None | Verified baseline |\n",
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "separator row with an extra cell",
			content: "| Pose | Mode | Image |\n" +
				"| --- | --- | ---: | --- |\n" +
				"| left | default | checked |\n",
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "body row with a missing cell",
			content: "| Result | Count | Notes |\n" +
				"| --- | --- | --- |\n" +
				"| pass | 1 |\n",
			wantViolations: 1,
			wantLine:       3,
		},
		{
			name: "separator row merged with a body row",
			content: "| Flag | Meaning |\n" +
				"| --- | --- | default |\n",
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "escaped pipe inside a cell",
			content: "| Pattern | Meaning |\n" +
				"| --- | --- |\n" +
				"| `a \\| b` | escaped pipe |\n",
			wantViolations: 0,
		},
		{
			name: "alignment markers",
			content: "| Result | Count | Notes |\n" +
				"| :----- | ----: | :---: |\n" +
				"| pass | 1 | first |\n",
			wantViolations: 0,
		},
		{
			name: "table inside a code fence",
			content: "```markdown\n" +
				"| Stage | Dependency | Deliverable |\n" +
				"| --- | --- |\n" +
				"| 00 | None | Verified baseline |\n" +
				"```\n",
			wantViolations: 0,
		},
		{
			name: "multiple tables",
			content: "| Stage | Dependency | Deliverable |\n" +
				"| --- | --- | --- |\n" +
				"| 00 | None | Verified baseline |\n" +
				"\n" +
				"| Result | Count | Notes |\n" +
				"| --- | --- |\n",
			wantViolations: 1,
			wantLine:       6,
		},
		{
			name: "carriage return line endings",
			content: "| Stage | Dependency | Deliverable |\r\n" +
				"| --- | --- |\r\n" +
				"| 00 | None | Verified baseline |\r\n",
			wantViolations: 1,
			wantLine:       2,
		},
		{
			name: "pipe block without a separator row",
			content: "| not | a | table |\n" +
				"| just | pipe | text |\n",
			wantViolations: 0,
		},
		{
			name: "indented pipe block",
			content: "    | Stage | Dependency | Deliverable |\n" +
				"    | --- | --- |\n",
			wantViolations: 0,
		},
	}

	rule := NewTableStructureRule()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := rule.CheckFile("table.md", []byte(tt.content))

			assert.Len(t, violations, tt.wantViolations)
			if tt.wantViolations == 0 {
				return
			}

			assert.Equal(t, tt.wantLine, violations[0].Position.Line)
			assert.Equal(t, "table.md", violations[0].Position.Filename)
			assert.Equal(t, rule.Name(), violations[0].Rule)
		})
	}
}
