ROOT := github.com/mfridman/tparse

.PHONY: vet
vet:
	@go vet ./...

.PHONY: lint
lint: tools
	@golangci-lint run ./... --fix

.PHONY: tools
tools:
	@cd tools && awk -F'"' '/^[[:space:]]*_[[:space:]]*"/ {print $$2}' tools.go | xargs -tI {} go install {}

.PHONY: test
test:
	go test -count=1 ./...

test-tparse:
	go test -race -count=1 ./internal/... -json -cover | go run main.go -trimpath=auto -sort=elapsed
	go test -race -count=1 ./tests/... -json -cover -coverpkg=./internal/parse | go run main.go -trimpath=github.com/mfridman/tparse/ -sort=elapsed

# dogfooding :)
test-tparse-full:
	go test -race -count=1 -v ./... -json | go run main.go -all -smallscreen -notests -sort=elapsed

.PHONY: release
release:
	goreleaser --rm-dist

coverage:
	go test ./parse -covermode=count -coverprofile=count.out
	go tool cover -html=count.out

.PHONY: build
build:
	go build -o $$GOBIN/tparse main.go

search-todo:
	@echo "Searching for TODOs in Go files..."
	@rg '// TODO\(mf\):' --glob '*.go' || echo "No TODOs found."

.PHONY: clean
clean:
	@find . -type f -name '*.FAIL' -delete
