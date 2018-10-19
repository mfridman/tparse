# tparse

A command line tool for analyzing and summarizing `go test` output.

<img src="https://raw.githubusercontent.com/mfridman/tparse/master/pass.png" width="50%" height="50%"/>

## Installation

    go get github.com/mfridman/tparse

## Usage

Once `tparse` is installed there are 2 ways to use it:

1. Run `go test` as you normally would, but add the `-json` flag and pipe the output to `tparse`.

Example:

```
go test fmt -json | tparse
```

2. Save the output of `go test` with the `-json` flag into a file and call `tparse` with filename as an argument.

```
go test fmt -json > fmt.out
tparse fmt.out
```