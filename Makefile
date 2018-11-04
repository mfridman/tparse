test:
	go test ./parse

# eating our own dog food :)
test-tparse-full:
	go test -count=1 -v ./... -json -cover | go run main.go -all -smallscreen -notests

test-parse:
	go test -count=1 ./parse -json -cover | go run main.go

release:
	goreleaser --rm-dist

coverage:
	go test ./parse -covermode=count -coverprofile=count.out
	go tool cover -html=count.out

generate:
	GIT_TAG=$$(git describe --tags) go generate ./...