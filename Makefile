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

test:
	go test -count=1 ./...

test-tparse:
	go test -race -count=1 ./... -json -cover -coverpkg=./... | go run main.go

# dogfooding :)
test-tparse-full:
	go test -race -count=1 -v ./... -json -cover -coverpkg=./... | go run main.go -all -smallscreen -notests

release:
	goreleaser --rm-dist

coverage:
	go test ./parse -covermode=count -coverprofile=count.out
	go tool cover -html=count.out

tidy:
	GO111MODULE=on go mod tidy && GO111MODULE=on go mod verify
