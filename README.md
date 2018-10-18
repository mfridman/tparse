# tparse

A cli tool for parsing the output of `go test` with `-json` flag. 

## Usage

    go get github.com/mfridman/tparse

Once tparse is installed:

1. run `go test` as you normally would, but add the `-json` flag and pipe the output into `tparse`.

Example:

```
go test fmt -json | tparse
```

2. save the output into a file and call `tparse` with filename as argument

```
go test fmt -json > fmt.out
tparse fmt.out
```