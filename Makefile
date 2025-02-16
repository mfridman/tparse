ROOT := github.com/mfridman/tparse
GOPATH ?= $(shell go env GOPATH)
TOOLS_BIN = $(GOPATH)/bin

.PHONY: vet
vet:
	@go vet ./...

.PHONY: lint
lint: tools
	@golangci-lint run ./... --fix

.PHONY: tools
tools:
	@which golangci-lint >/dev/null 2>&1 || \
		(echo "Installing latest golangci-lint" && \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b "$(TOOLS_BIN)")

.PHONY: tools-update
tools-update:
	@echo "Updating golangci-lint to latest version"
	@rm -f "$(TOOLS_BIN)/golangci-lint"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b "$(TOOLS_BIN)"
	@echo "golangci-lint updated successfully to latest version"

.PHONY: tools-version
tools-version:
	@echo "Current tool versions:"
	@echo "golangci-lint: $$(golangci-lint --version 2>/dev/null || echo 'not installed')"

.PHONY: release
release:
	@goreleaser --rm-dist

.PHONY: build
build:
	@go build -o $$GOBIN/tparse ./

.PHONY: clean
clean:
	@find . -type f -name '*.FAIL' -delete

.PHONY: test
test:
	@go test -count=1 ./...

test-tparse:
	@go test -race -count=1 ./internal/... -json -cover | go run main.go -trimpath=auto -sort=elapsed
	@go test -race -count=1 ./tests/... -json -cover -coverpkg=./parse | go run main.go -trimpath=github.com/mfridman/tparse/ -sort=elapsed

# dogfooding :)
test-tparse-full:
	go test -race -count=1 -v ./... -json | go run main.go -all -smallscreen -notests -sort=elapsed

coverage:
	go test ./tests/... -coverpkg=./parse -covermode=count -coverprofile=count.out
	go tool cover -html=count.out

search-todo:
	@echo "Searching for TODOs in Go files..."
	@rg '// TODO\(mf\):' --glob '*.go' || echo "No TODOs found."

