# retrogolint - a linter for retrogolib projects

[![Build status](https://github.com/retroenv/retrogolint/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/retroenv/retrogolint/actions)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/retroenv/retrogolint)
[![Go Report Card](https://goreportcard.com/badge/github.com/retroenv/retrogolint)](https://goreportcard.com/report/github.com/retroenv/retrogolint)

## Installation

```bash
go install github.com/retroenv/retrogolint/cmd/retrogolint@latest
```

**Requirements:**
- Go 1.22 or later

## Overview

Retrogolint is a Go static analyzer for projects built around [`retrogolib`](https://github.com/retroenv/retrogolib).
It checks retrogolib-specific logging conventions, testing style, collection usage, and code organization rules that are not covered by general-purpose Go linters.

### Key Design Principles

- **retrogolib-aware rules**: Encodes conventions for retrogolib logging, assertions, collections, and package structure
- **AST-based analysis**: Uses Go's standard parser and AST model for deterministic checks
- **CI-friendly output**: Supports concise text output and structured JSON output
- **Configurable scope**: Enables rule/category selection, global exclusions, and per-rule exclusions
- **Small command-line tool**: Builds as a static CLI with no runtime service dependencies

## Rule Coverage

### Current Support

- **Logging**: Message capitalization, field casing, nil checks, zero-value loggers, formatting helpers, specialized fields, and eager field evaluation
- **Testing**: Manual assertion patterns that should use `retrogolib/assert`
- **Collections**: Set-like `map[K]bool` and `map[K]struct{}` patterns that should use `retrogolib/set`
- **Code quality**: Top-level declaration order, parameter priority, and exported type stutter

See [docs/rules.md](docs/rules.md) for the complete rule list.

## Features

### Command-Line Analysis

- Analyze files, directories, and Go package patterns such as `./...`
- Select rules by exact rule name or by category
- Filter reported violations by minimum severity
- Exclude test files, directories, and filename or relative path patterns

### Configuration

- Read `.retrogolint.ini` by default
- Override file settings with command-line flags
- Configure global exclusions and per-rule exclusions
- See [docs/configuration.md](docs/configuration.md) for the full configuration reference

## Package Overview

    ├─ cmd/retrogolint      command-line entry point
    ├─ docs                 configuration, rules, and development reference
    ├─ internal/analyzer    package/file expansion and parallel AST analysis
    ├─ internal/cli         command-line flag parsing and config merging
    ├─ internal/linterconfig configuration loading and exclusion rules
    ├─ internal/reporter    text and JSON output formatting
    ├─ internal/rules       built-in lint rule registry and implementations
    ├─ internal/violation   shared violation and severity model
    ├─ internal/workflow    high-level analysis orchestration
    ├─ testdata             valid and invalid rule fixtures

## Quick Start

### CLI Usage

Analyze the current module:

```bash
retrogolint ./...
```

Analyze a package tree:

```bash
retrogolint ./internal/compiler/...
```

Output JSON for CI or tooling:

```bash
retrogolint -format json ./...
```

Run only logging rules:

```bash
retrogolint -rules logging ./...
```

Disable a rule or category:

```bash
retrogolint -disabled-rules collections-map-set ./...
```

Show command usage:

```text
usage: retrogolint [options] [packages...]

Parameters:
  -config string
        Configuration file path (default: .retrogolint.ini)

Options:
  -severity string
        Minimum severity to report: error, warning, info (default: warning)
  -rules string
        Comma-separated list of rule categories to check (default: all rules)
  -disabled-rules string
        Comma-separated list of rules/categories to disable
  -max-per-rule int
        Maximum violations per rule (0 = unlimited)
  -exclude-tests
        Exclude test files from analysis
  -exclude-dirs string
        Comma-separated list of directories to exclude
  -exclude-files string
        Comma-separated list of filename or relative path patterns to exclude
  -version
        Show version information

Output options:
  -format string
        Output format: text, json (default: text)

Positional arguments:
  packages...
        packages to analyze
```

If no paths are provided, `retrogolint` analyzes `./...`.

## Development

```bash
make build
make lint
make test
```

See [docs/development.md](docs/development.md) for release checks and local development commands.

## License

This project is licensed under the Apache License Version 2.0 - see the [LICENSE](LICENSE) file for details.
