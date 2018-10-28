test:
	go test ./tests ./parse

# eating our own dog food :)
test-tparse:
	go test -count=1 -v ./... -json -cover | go run main.go -all -smallscreen -notests

release:
	goreleaser --rm-dist