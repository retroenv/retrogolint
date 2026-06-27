# Changelog for retrogolint

All notable changes to this project will be documented in this file.

## [v1.0.3] - 2026-0x-xx

Changed:

* logging rules now analyze wrapper calls that pass structured `log.*` fields, including message capitalization, field casing, formatting, hex casts, specialized fields, eager field evaluation, and `.String()` method checks.
* shared log-call detection now tracks message and field argument positions for direct logger calls and wrapper calls.

Fixed:

* sectioned `.retrogolint.ini` files now apply global settings from `[output]`, `[rules]`, and `[integration]`, so options such as `severity = info` are honored without requiring matching CLI flags.
* `logging-efficiency` no longer reports built-in calls such as `len(...)` and `cap(...)`, which cannot use lazy field helpers directly and produced noisy false positives.

## [v1.0.2] - 2026-05-03

Changed:

* `codequality-funcorder` now enforces dependency ordering for exported types:
  unexported dependency types must be declared before the exported type that uses them.
* dependency checks cover both exported type definitions and method signatures on exported receiver types.
* declaration ordering allows unexported dependency types directly before the exported type they support.

Fixed:

* `codequality-funcorder` false positives where dependency types were placed correctly before exported types.

## [v1.0.1] - 2026-05-03

Added:

* add relative path support for file exclusions

Fixed:

* fix goreleaser deprecation warnings

## [v1.0.0] - 2026-05-03

First version of retrogolint released.
