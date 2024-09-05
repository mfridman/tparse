ROOT := github.com/mfridman/tparse

.PHONY: \
	imports \
	test \
	tidy \

check: test vet staticcheck imports

vet:
	go vet ./...

staticcheck:
	@which staticcheck 2>/dev/null || go get -u honnef.co/go/tools/cmd/staticcheck
	staticcheck $(ROOT) ./parse

errcheck:
	@errcheck -help 2>/dev/null || go get -u github.com/kisielk/errcheck
	errcheck $(PKGS)

imports:
	goimports -local $(ROOT) -w $(shell find . -type f -name '*.go' -not -path './vendor/*')

.PHONY: lint
lint: tools
	@golangci-lint run ./... --fix

.PHONY: tools
tools:
# Install latest golangci-lint with recommended method https://golangci-lint.run/welcome/install/#local-installation
# Only install it if missing, as we don't want to mess up with any local existing golangci-lint version
	@which golangci-lint 2>/dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

test:
	go test -count=1 ./...

test-tparse:
	go test -race -count=1 ./internal/... -json -cover | go run main.go -trimpath=auto -sort=elapsed
	go test -race -count=1 ./tests/... -json -cover -coverpkg=./parse | go run main.go -trimpath=github.com/mfridman/tparse/ -sort=elapsed

# dogfooding :)
test-tparse-full:
	go test -race -count=1 -v ./... -json | go run main.go -all -smallscreen -notests -sort=elapsed

release:
	goreleaser --rm-dist

coverage:
	go test ./parse -covermode=count -coverprofile=count.out
	go tool cover -html=count.out

tidy:
	GO111MODULE=on go mod tidy && GO111MODULE=on go mod verify

build:
	go build -o $$GOBIN/tparse main.go

search-todo:
	@echo "Searching for TODOs in Go files..."
	@rg '// TODO\(mf\):' --glob '*.go' || echo "No TODOs found."

.PHONY: clean
clean:
	@find . -type f -name '*.FAIL' -delete
