# to get a specific version from https://golang.org/dl/
# go get golang.org/dl/go1.11.4 && go1.11.4 download
# 
GO_VERSION=go1.11.4
# GO_VERSION=go1.12beta1
# GO_VERSION=gotip
GO=$$HOME/go/bin/$(GO_VERSION)

ROOT := github.com/mfridman/tparse

vet: | test
	$(GO) vet $(PKGS)

go-version:
	$(GO) version

.PHONY: \
	imports \
	test \
	tidy \
	vendor \

imports:
	@goimports -local $(ROOT) -w $(shell find . -type f -name '*.go' -not -path './vendor/*')

test:
	$(GO) test -count=1 ./parse

test-tparse:
	$(GO) test -race -count=1 ./parse -json -cover | $(GO) run main.go

# eating our own dog food :)
test-tparse-full:
	$(GO) test -race-count=1 -v ./... -json -cover | $(GO) run main.go -all -smallscreen -notests

release:
	goreleaser --rm-dist

coverage:
	$(GO) test ./parse -covermode=count -coverprofile=count.out
	$(GO) tool cover -html=count.out

tidy:
	GO111MODULE=on $(GO) mod tidy && GO111MODULE=on $(GO) mod verify

update-patch:
	GO111MODULE=on $(GO) get -u=patch

generate:
	GIT_TAG=$$(git describe --tags) $(GO) generate ./...
