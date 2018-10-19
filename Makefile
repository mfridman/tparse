test:
	go test -v ./tests

# eating our own dog food :)
test-tparse:
	go test -count=1 -v ./tests -json | go run main.go 