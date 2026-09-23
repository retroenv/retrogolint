// Package markdown holds lint rules for markdown documents.
package markdown

import (
	"fmt"
	"go/token"
	"regexp"
	"strings"

	"github.com/retroenv/retrogolint/internal/rules/api"
	"github.com/retroenv/retrogolint/internal/violation"
)

const (
	escapedPipePlaceholder = "\x00"
	tableIndentLimit       = 3
)

// separatorCellPattern matches a markdown table separator cell, such as "---" or ":--:".
var separatorCellPattern = regexp.MustCompile(`^:?-+:?$`)

// TableStructureRule detects markdown tables with inconsistent cell counts.
type TableStructureRule struct{}

// NewTableStructureRule creates a new TableStructureRule.
func NewTableStructureRule() *TableStructureRule {
	return &TableStructureRule{}
}

// Name returns the rule name.
func (r *TableStructureRule) Name() string {
	return "markdown-table-structure"
}

// Description returns the rule description.
func (r *TableStructureRule) Description() string {
	return "Markdown tables use the same cell count in the header, separator, and body rows"
}

// Severity returns the default severity.
func (r *TableStructureRule) Severity() violation.Severity {
	return violation.SeverityWarning
}

// Category returns the rule category.
func (r *TableStructureRule) Category() string {
	return api.CategoryMarkdown
}

// CheckFile analyzes a markdown document for malformed tables.
func (r *TableStructureRule) CheckFile(path string, content []byte) []violation.Violation {
	lines := splitLines(content)

	var violations []violation.Violation
	var fence string

	for i := 0; i < len(lines); {
		if marker := fenceMarker(lines[i]); marker != "" {
			fence = toggleFence(fence, marker)
			i++
			continue
		}

		if fence != "" || !isTableRow(lines[i]) {
			i++
			continue
		}

		start := i
		for i < len(lines) && isTableRow(lines[i]) {
			i++
		}

		violations = append(violations, r.checkBlock(path, start, lines[start:i])...)
	}

	return violations
}

// checkBlock verifies one table block that starts at the given zero-based line index.
func (r *TableStructureRule) checkBlock(path string, start int, block []string) []violation.Violation {
	if len(block) < 2 {
		return nil
	}

	headerCells := len(splitTableCells(block[0]))
	separatorCells := splitTableCells(block[1])

	if !isSeparatorRow(separatorCells) {
		if isMixedSeparatorRow(separatorCells) {
			found := r.newViolation(path, start+2, 1,
				"table separator row mixes separator and content cells")
			return []violation.Violation{found}
		}
		return nil
	}

	var violations []violation.Violation
	if len(separatorCells) != headerCells {
		message := fmt.Sprintf("table separator row has %d cells but the header has %d",
			len(separatorCells), headerCells)
		violations = append(violations, r.newViolation(path, start+2, 1, message))
	}

	for offset := 2; offset < len(block); offset++ {
		count := len(splitTableCells(block[offset]))
		if count == headerCells {
			continue
		}

		message := fmt.Sprintf("table row has %d cells but the header has %d", count, headerCells)
		violations = append(violations, r.newViolation(path, start+offset+1, 1, message))
	}

	return violations
}

func (r *TableStructureRule) newViolation(path string, line, column int, message string) violation.Violation {
	position := token.Position{
		Filename: path,
		Line:     line,
		Column:   column,
	}

	return violation.Violation{
		Rule:     r.Name(),
		Message:  message,
		Position: position,
		Severity: r.Severity(),
	}
}

// splitLines normalizes line endings and splits the content into lines.
func splitLines(content []byte) []string {
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	return strings.Split(normalized, "\n")
}

// fenceMarker returns the code fence marker that starts the line, or an empty string.
func fenceMarker(line string) string {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > tableIndentLimit {
		return ""
	}

	if strings.HasPrefix(trimmed, "```") {
		return "```"
	}
	if strings.HasPrefix(trimmed, "~~~") {
		return "~~~"
	}

	return ""
}

// toggleFence opens a code fence or closes the fence that matches the marker.
func toggleFence(fence, marker string) string {
	switch fence {
	case "":
		return marker
	case marker:
		return ""
	default:
		return fence
	}
}

// isTableRow reports whether the line can start a markdown table row.
func isTableRow(line string) bool {
	trimmed := strings.TrimLeft(line, " ")
	if len(line)-len(trimmed) > tableIndentLimit {
		return false
	}

	return strings.HasPrefix(trimmed, "|")
}

// splitTableCells splits a table row into its trimmed cells.
func splitTableCells(line string) []string {
	normalized := strings.ReplaceAll(line, `\|`, escapedPipePlaceholder)
	normalized = strings.TrimSpace(normalized)
	normalized = strings.TrimPrefix(normalized, "|")
	normalized = strings.TrimSuffix(normalized, "|")

	parts := strings.Split(normalized, "|")
	cells := make([]string, 0, len(parts))

	for _, part := range parts {
		cell := strings.ReplaceAll(strings.TrimSpace(part), escapedPipePlaceholder, "|")
		cells = append(cells, cell)
	}

	return cells
}

// isSeparatorRow reports whether every cell is a table separator cell.
func isSeparatorRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}

	for _, cell := range cells {
		if !isSeparatorCell(cell) {
			return false
		}
	}

	return true
}

// isMixedSeparatorRow reports whether the row mixes separator cells with content cells.
func isMixedSeparatorRow(cells []string) bool {
	hasSeparator := false
	hasContent := false

	for _, cell := range cells {
		switch {
		case isSeparatorCell(cell):
			hasSeparator = true
		case cell != "":
			hasContent = true
		}
	}

	return hasSeparator && hasContent
}

func isSeparatorCell(cell string) bool {
	return separatorCellPattern.MatchString(cell)
}
