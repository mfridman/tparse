# tparse  [![Build Status](https://travis-ci.com/mfridman/tparse.svg?branch=master)](https://travis-ci.com/mfridman/tparse)

A command line tool for analyzing and summarizing `go test` output.

**Don't forget to run `go test` with `-json` flag.**

Pass            |  Fail
:-------------------------:|:-------------------------:
<img src="https://www.dropbox.com/s/tx7hod8lf646qgw/pass.png?raw=1" />  |  <img src="https://www.dropbox.com/s/d5bzagnjewcf338/fail.png?raw=1" />

By default, `tparse` will always return a summary box containing package-level details followed by errors, if any.

To get the summary of passed tests run `tparse` with the `-pass` flag. Tests will be grouped by package and sorted by elapsed time (longest to shorted).

See [Extras](#extras) for more info.

## Installation

    go get github.com/mfridman/tparse

## Usage

Once `tparse` is installed there are 2 ways to use it:

1. Run `go test` as you normally would, but add the `-json` flag and pipe the output to `tparse`.

```
go test fmt -json | tparse -all
```

2. Save the output of `go test` with the `-json` flag into a file and call `tparse` with filename as an argument.

```
go test fmt -json > fmt.out
tparse -all fmt.out
```

Tip: run `tparse -h` to get usage and options.

## Extras

`go test` is a great tool, but a bit verbose. Sometimes all one wants is failures grouped by package printed with a dash of color and bubbled to the top of the output.

By default, `tparse` does just that, outputs a summary box and all failed tests (if any).

But we can take it a bit further. With `-all` (`-pass` and `-skip` combined) we can get additional information, such as which tests were skipped or the elapsed time of each passed test.

`tparse` comes with the `-dump` flag to print back everything that was read. For those wanting to retrieve the original `go test` output.

Some CI pipelines and terminals are narrow, so give `-smallscreen` flag a try. It takes long test names and makes them vertical heavy:

```
TestSubtests/an_awesome_but_long/subtest_for_the/win

TestSubtests
 /an_awesome_but_long
 /subtest_for_the
 /win
 ```

So, one-liner bashers (were|are) fun, but `tparse` aims to provide a simply alternative.

p.s. `tparse` uses itself in its travis pipeline:

Example [here](https://travis-ci.com/mfridman/tparse/jobs/154695634)

<img src="https://www.dropbox.com/s/4tq8m8dhjphn7b7/travis-ci.png?raw=1" />