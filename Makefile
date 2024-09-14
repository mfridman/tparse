ROOT := github.com/mfridman/tparse

.PHONY: vet
vet:
	@go vet ./...

.PHONY: lint
lint: tools
	@golangci-lint run ./... --fix

.PHONY: tools
tools:
	@which golangci-lint >/dev/null 2>&1 || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

.PHONY: release
release:
	@goreleaser --rm-dist

.PHONY: build
build:
	@go build -o $$GOBIN/tparse main.go

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

