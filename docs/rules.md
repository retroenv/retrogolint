# Rules

Rules can be selected by exact rule name or by category. Built-in categories are `codequality`, `collections`, `logging`, and `testing`.

## Code Quality

| Rule | Severity | Checks |
|------|----------|--------|
| `codequality-funcorder` | warning | Top-level declarations follow order: exported types, constructors, methods → unexported types, constructors, methods → functions |
| `codequality-param-priority` | warning | Function parameters put `context.Context` first, then logger parameters, then other parameters |
| `codequality-type-stutter` | warning | Exported type names do not repeat the package name |

`codequality-funcorder` ignores `init` functions and `_test.go` files.

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
