# Configuration

`retrogolint` reads `.retrogolint.ini` by default. Pass `-config` to use another path. Command-line flags override configuration file values.

## Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `.retrogolint.ini` | Configuration file path |
| `-format` | `text` | Output format: `text`, `json` |
| `-severity` | `warning` | Minimum severity: `error`, `warning`, `info` |
| `-rules` | all rules | Comma-separated rule categories or rule names to run |
| `-disabled-rules` | none | Comma-separated rule categories or rule names to disable |
| `-max-per-rule` | `0` | Maximum violations per rule; `0` means unlimited |
| `-exclude-tests` | `false` | Exclude `_test.go` files |
| `-exclude-dirs` | none | Comma-separated directory names to exclude |
| `-exclude-files` | none | Comma-separated filename patterns to exclude |
| `-version` | `false` | Print version information |

## File Example

```ini
format = json
severity = warning
rules =
disabled-rules = logging-efficiency
max-per-rule = 10
exclude-tests = true
exclude-dirs = vendor,testdata
exclude-files = *_generated.go
```

Notes:

- Keys may use `-` or `_` separators.
- Lines starting with `#` or `;` are comments.
- Empty `rules` means all rules are enabled.
- Unknown rule names or categories in `rules` and `disabled-rules` are reported as errors.

## Per-Rule Exclusions

Per-rule exclusions use `[rule.*]` sections. They are combined with global exclusions.

```ini
[rule.logging-field-casing]
exclude-files = *_generated.go,*_mock.go
exclude-dirs = testdata

[rule.logging-efficiency]
exclude-tests = true

[rule.logging]
exclude-files = legacy_*.go
```

Rule-level sections apply to one rule. Category sections, such as `[rule.logging]`, apply to every rule in that category.

## Output Formats

Text output is intended for humans:

```text
main.go:15:2: Log message should start with an uppercase letter (logging-capitalization)

Found 1 violation(s)
```

JSON output is intended for tooling:

```json
{
  "violations": [
    {
      "rule": "logging-capitalization",
      "message": "Log message should start with an uppercase letter",
      "position": {
        "filename": "main.go",
        "line": 15,
        "column": 2
      },
      "severity": "warning"
    }
  ],
  "count": 1
}
```
