// Package violation defines types for representing linting violations.
package violation

import (
	"encoding/json"
	"fmt"
	"go/token"
)

// Severity represents the severity level of a violation.
type Severity int

// Violation represents a linting violation found in the code.
type Violation struct {
	// Rule is the name of the rule that detected this violation.
	Rule string `json:"rule"`

	// Message is a human-readable description of the violation.
	Message string `json:"message"`

	// Position is the location in the source file where the violation occurs.
	Position token.Position `json:"position"`

	// Severity indicates how serious the violation is.
	Severity Severity `json:"severity"`

	// Context provides surrounding code context for the violation.
	Context string `json:"context,omitempty"`
}

// ShouldReport returns true if the violation should be reported given the minimum severity.
func (v *Violation) ShouldReport(minSeverity Severity) bool {
	return v.Severity >= minSeverity
}

const (
	// SeverityInfo represents informational messages.
	SeverityInfo Severity = iota
	// SeverityWarning represents warnings.
	SeverityWarning
	// SeverityError represents errors.
	SeverityError
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// MarshalJSON implements json.Marshaler.
func (s Severity) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(s.String())
	if err != nil {
		return nil, fmt.Errorf("failed to marshal severity: %w", err)
	}
	return data, nil
}

// ParseSeverity parses a string into a Severity value.
func ParseSeverity(s string) Severity {
	switch s {
	case "info":
		return SeverityInfo
	case "warning":
		return SeverityWarning
	case "error":
		return SeverityError
	default:
		return SeverityWarning
	}
}
