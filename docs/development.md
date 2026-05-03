# Development

## Common Commands

```bash
make build
make lint
make test
make test-coverage
make install
make clean
```

`make build` creates a local `retrogolint` binary from `./cmd/retrogolint`.

## Release Checks

Before tagging a release, run:

```bash
go fmt ./...
make lint
make test
go vet ./...
go run ./cmd/retrogolint ./...
```

The repository includes `.retrogolint.ini` to exclude intentional fixtures in `testdata`.

## Versioning

The version, commit, and build date are injected at build time via ldflags in the `Makefile`:

```
-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)
```

`VERSION` defaults to the output of `git describe --tags --always --dirty`. Print the version with:

```bash
retrogolint -version
```
