# to get a specific version from https://golang.org/dl/
# go get golang.org/dl/go1.11.4 && go1.11.4 download
# 
# GO_VERSION := go1.11.4
# GO_VERSION := go1.12beta1
# GO_VERSION := gotip
# GO := $$HOME/go/bin/$(GO_VERSION)
GO := go

ROOT := github.com/mfridman/tparse

.PHONY: \
	imports \
	test \
	tidy \

go-version:
	$(GO) version

check: test vet staticcheck imports

vet:
	$(GO) vet ./...

staticcheck:
	@which staticcheck 2>/dev/null || $(GO) get -u honnef.co/go/tools/cmd/staticcheck
	staticcheck $(ROOT) ./parse

errcheck:
	@errcheck -help 2>/dev/null || $(GO) get -u github.com/kisielk/errcheck
	errcheck $(PKGS)

imports:
	goimports -local $(ROOT) -w $(shell find . -type f -name '*.go' -not -path './vendor/*')

test:
	$(GO) test -count=1 ./parse

test-tparse:
	$(GO) test -race -count=1 ./parse -json -cover | $(GO) run main.go

# dogfooding :)
test-tparse-full:
	$(GO) test -race -count=1 -v ./... -json -cover | $(GO) run main.go -all -smallscreen -notests

release:
	goreleaser --rm-dist

coverage:
	$(GO) test ./parse -covermode=count -coverprofile=count.out
	$(GO) tool cover -html=count.out

tidy:
	GO111MODULE=on $(GO) mod tidy && GO111MODULE=on $(GO) mod verify

update-patch:
	GO111MODULE=on $(GO) get -u=patch
