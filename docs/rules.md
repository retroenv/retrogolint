# Rules

Rules can be selected by exact rule name or by category. Built-in categories are `codequality`, `collections`, `logging`, `markdown`, and `testing`.

## Code Quality

| Rule | Severity | Checks |
|------|----------|--------|
| `codequality-func-signature-layout` | warning | Function and method signatures stay on one line through 120 columns; longer signatures keep parameter names on the earliest line where they fit |
| `codequality-funcorder` | warning | Top-level declarations follow order: exported types, constructors, methods → unexported types, constructors, methods → functions; unexported dependency types must be declared before exported types that use them |
| `codequality-param-priority` | warning | Function parameters put `context.Context` first, then logger parameters, then other parameters |
| `codequality-param-type-combine` | warning | Consecutive function parameters with the same type share one type declaration |
| `codequality-struct-literal-multiline` | warning | Keyed struct literals with two or more fields use one line per field |
| `codequality-type-stutter` | warning | Exported type names do not repeat the package name |

`codequality-funcorder` ignores `init` functions and `_test.go` files.
Dependency usage is detected in exported type definitions and method signatures on exported receiver types.

Retrogolint provides `codequality-param-type-combine` because GoCritic's `paramTypeCombine` check skips multiline parameter lists. Long signatures need the rule when line wrapping makes repeated types less visible.

Example:

```go
type dep interface{ Run() }

type Service struct {
	dep dep
}

func (s *Service) Tick(m dep) {}
```

## Collections

| Rule | Severity | Checks |
|------|----------|--------|
| `collections-map-set` | warning | `map[K]bool` or `map[K]struct{}` used as a set |

Prefer `github.com/retroenv/retrogolib/set.Set` for explicit set semantics.

## Logging

| Rule | Severity | Checks |
|------|----------|--------|
| `logging-capitalization` | warning | Log messages start with an uppercase letter |
| `logging-efficiency` | info | Eager method calls in log fields that should be lazy |
| `logging-field-casing` | warning | Log field names use `snake_case` |
| `logging-fmt-in-field` | warning | `fmt` calls inside log field values |
| `logging-hex-cast` | warning | Unnecessary widening casts before `log.Hex()` |
| `logging-hex-formatting` | warning | `fmt.Sprintf` hex formatting where `log.Hex()` is preferred |
| `logging-method-calls` | warning | `log.String(k, v.String())` calls that should use `log.Stringer()` |
| `logging-nil-check` | warning | Unnecessary nil checks before logger calls |
| `logging-specialized-field` | warning | Generic integer fields used after casting values with narrower types |
| `logging-type-formatting` | warning | `fmt.Sprintf("%T", ...)` where `log.Type()` is preferred |
| `logging-zero-value` | warning | Unsafe `&log.Logger{}` zero-value construction |

Examples:

```go
logger.Info("starting server") // logging-capitalization
logger.Info("Load", log.String("fileName", name)) // logging-field-casing
logger.Debug("Handler", log.String("addr", fmt.Sprintf("0x%04X", address))) // logging-hex-formatting
```

Prefer:

```go
logger.Info("Starting server")
logger.Info("Load", log.String("file_name", name))
logger.Debug("Handler", log.Hex("addr", address))
```

## Markdown

| Rule | Severity | Checks |
|------|----------|--------|
| `markdown-table-structure` | warning | Table headers, separator rows, and body rows use the same cell count |

A table renders incorrectly when the separator row or a body row has a different
cell count than the header. The rule ignores tables inside fenced code blocks.

Example:

```markdown
| Stage | Dependency | Deliverable |
| --- | --- |
| 00 | None | Verified baseline |
```

Prefer:

```markdown
| Stage | Dependency | Deliverable |
| --- | --- | --- |
| 00 | None | Verified baseline |
```

## Testing

| Rule | Severity | Checks |
|------|----------|--------|
| `assert-usage` | warning | Manual test assertions that should use `retrogolib/assert` helpers |

Examples:

```go
if err != nil {
    t.Fatal(err)
}

if len(items) != 3 {
    t.Fatalf("got %d", len(items))
}
```

Prefer:

```go
assert.NoError(t, err)
assert.Len(t, items, 3)
```
