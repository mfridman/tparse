# tparse  [![Actions](https://github.com/mfridman/tparse/workflows/CI/badge.svg)](https://github.com/mfridman/tparse)

A command line tool for analyzing and summarizing `go test` output.

**Don't forget to run `go test` with the `-json` flag.**

Pass            |  Fail
:-------------------------:|:-------------------------:
<img src="https://user-images.githubusercontent.com/6278244/170038081-1ddc5486-7c97-49a3-ac2d-08b502e39bdf.png" />  |  <img src="https://user-images.githubusercontent.com/6278244/170038118-3cecdb30-411c-4534-84b3-0a55db85cb1e.png" />

By default, `tparse` will always return test failures and panics, if any, followed by a package-level summary table.

To get additional info on passed tests run `tparse` with `-pass` flag. Tests are grouped by package and sorted by elapsed time in descending order (longest to shortest).

### [But why?!](#but-why) for more info.

## Installation

    go install github.com/mfridman/tparse@latest

Or download the latest pre-built binary [here](https://github.com/mfridman/tparse/releases/latest).

## Usage

Once `tparse` is installed there are 2 ways to use it:

1. Run `go test` as normal, but add `-json` flag and pipe output to `tparse`.

```
set -o pipefail && go test fmt -json | tparse -all
```

2. Save the output of `go test` with `-json` flag into a file and call `tparse` with `-file` option.

```
go test fmt -json > fmt.out
tparse -all -file=fmt.out
```

Tip: run `tparse -h` to get usage and options.

## But why?!

`go test` is awesome, but verbose. Sometimes you just want readily available failures, grouped by package, printed with a dash of color.

`tparse` attempts to do just that; return failed tests and panics, if any, followed by a single package-level summary. No more searching for the literal string: "--- FAIL".

But, let's take it a bit further. With `-all` (`-pass` and `-skip` combined) you can get additional info, such as skipped tests and elapsed time of each passed test.

`tparse` comes with a `-follow` flag to print raw output. Yep, go test pipes JSON, it's parsed and the output is printed back out as if you ran go test without `-json` flag. Eliminating the need for `tee /dev/tty` between pipes.

The default print order is:
- `go test` output (if adding `-follow` flag)
- passed/skipped table (if adding `-all`, `-skip` or `-pass` flag)
- failed tests and panics
- summary

For narrow displays the `-smallscreen` flag may be useful, dividing a long test name and making it vertical heavy:

```
TestSubtests/an_awesome_but_long/subtest_for_the/win

TestSubtests
 /an_awesome_but_long
 /subtest_for_the
 /win
 ```

`tparse` aims to be a simply alternative to one-liner bash functions.
