GOLANGCI_VERSION = v2.12.1

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
DATE    ?= $(shell git log -1 --format=%cI 2>/dev/null || echo "")
LDFLAGS  = -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

help: ## show help, shown by default if no target is specified
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

lint: ## run code linters
	golangci-lint run

build: ## build retrogolint binary
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o retrogolint ./cmd/retrogolint

install: ## install retrogolint binary
	CGO_ENABLED=0 go install -ldflags "$(LDFLAGS)" ./cmd/retrogolint

test: ## run tests
	go test -timeout 10s -race ./...

test-coverage: ## run unit tests and create test coverage
	go test -timeout 10s ./... -coverprofile coverage.txt

test-coverage-web: test-coverage ## run unit tests and show test coverage in browser
	go tool cover -func coverage.txt | grep total | awk '{print "Total coverage: "$$3}'
	go tool cover -html=coverage.txt

install-linters: ## install all used linters
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_VERSION}

clean: ## remove local build and coverage artifacts
	rm -f retrogolint coverage.txt
