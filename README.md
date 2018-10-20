# tparse  [![Build Status](https://travis-ci.com/mfridman/tparse.svg?branch=master)](https://travis-ci.com/mfridman/tparse)

A command line tool for analyzing and summarizing `go test` output.

Pass            |  Fail
:-------------------------:|:-------------------------:
<img src="https://www.dropbox.com/s/255g8pgl55unes9/tparse_pass.png?raw=1" />  |  <img src="https://www.dropbox.com/s/x3nt5nztv3em94o/tparse_fail.png?raw=1" />

## Installation

    go get github.com/mfridman/tparse

## Usage

Once `tparse` is installed there are 2 ways to use it:

1. Run `go test` as you normally would, but add the `-json` flag and pipe the output to `tparse`.

Example:

```
go test fmt -json | tparse -all
```

2. Save the output of `go test` with the `-json` flag into a file and call `tparse` with filename as an argument.

```
go test fmt -json > fmt.out
tparse -all fmt.out
```