# retrogolint

[![CI](https://github.com/retroenv/retrogolint/actions/workflows/go.yaml/badge.svg?branch=main)](https://github.com/retroenv/retrogolint/actions/workflows/go.yaml)
[![Codecov](https://codecov.io/gh/retroenv/retrogolint/graph/badge.svg)](https://codecov.io/gh/retroenv/retrogolint)
[![Release](https://img.shields.io/github/v/release/retroenv/retrogolint)](https://github.com/retroenv/retrogolint/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/retroenv/retrogolint.svg)](https://pkg.go.dev/github.com/retroenv/retrogolint)
[![License](https://img.shields.io/github/license/retroenv/retrogolint)](LICENSE)
![LLM assisted: human reviewed](https://img.shields.io/badge/LLM%20assisted-human%20reviewed-6f42c1)

A Go static analyzer that enforces conventions for projects built with
[`retrogolib`](https://github.com/retroenv/retrogolib).

## Features

* **Retrogolib-aware checks** - Finds issues that general-purpose Go linters do not cover
* **Flexible targets** - Analyzes files, directories, and Go package patterns such as `./...`
* **Rule selection** - Runs or disables individual rules and complete rule categories
* **Configurable severity** - Reports violations at or above the selected severity
* **Targeted exclusions** - Excludes tests, directories, and file patterns globally or per rule
* **CI output** - Produces concise text or structured JSON output

## Rule Categories

| Category | Checks |
|----------|--------|
| **Logging** | Message style, structured fields, logger use, and eager field evaluation |
| **Testing** | Manual assertions that can use `retrogolib/assert` |
| **Collections** | Map-based sets and inefficient `retrogolib/set` operations |
| **Code quality** | Declaration order, parameter order, and exported type names |

See the [rule reference](docs/rules.md) for rule details and examples.

## Quick Start

### Installation

Download a prebuilt binary from
[Releases](https://github.com/retroenv/retrogolint/releases), or install from
source with Go 1.22 or later:

```bash
go install github.com/retroenv/retrogolint/cmd/retrogolint@latest
```

### Basic Usage

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

If no paths are provided, `retrogolint` analyzes `./...`.

## Configuration

`retrogolint` reads `.retrogolint.ini` by default. Command-line flags override
file settings. The configuration supports global exclusions and exclusions for
individual rules or rule categories.

See the [configuration reference](docs/configuration.md) for all options and
examples.
