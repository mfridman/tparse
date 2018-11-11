# tparse  [![Build Status](https://travis-ci.com/mfridman/tparse.svg?branch=master)](https://travis-ci.com/mfridman/tparse)

A command line tool for analyzing and summarizing `go test` output.

**Don't forget to run `go test` with the `-json` flag.**

Pass            |  Fail
:-------------------------:|:-------------------------:
<img src="https://www.dropbox.com/s/c6v4u9d0huvjelk/pass.png?raw=1" />  |  <img src="https://www.dropbox.com/s/k4lgoekqef7tbsy/fail.png?raw=1" />

By default, `tparse` will always return test failures and panics, if any, followed by a package-level summary table.

To get additional info on passed tests run `tparse` with `-pass` flag. Tests are grouped by package and sorted by elapsed time in descending order (longest to shortest).

### [But why?!](#but-why) for more info.

## Installation

    go get github.com/mfridman/tparse

## Usage

Once `tparse` is installed there are 2 ways to use it:

1. Run `go test` as you normally would, but add `-json` flag and pipe output to `tparse`.

```
go test fmt -json | tparse -all
```

2. Save the output of `go test` with `-json` flag into a file and call `tparse` with filename as an argument.

```
go test fmt -json > fmt.out
tparse -all fmt.out
```

Tip: run `tparse -h` to get usage and options.

## But why?!

`go test` is awesome, but a bit verbose. Sometimes one just wants failures, grouped by package, printed with a dash of color and bubbled to the top.

`tparse` attempts to do just that; return all failed tests and panics, if any, followed by a single package-level summary.

But, let's take it a bit further. With `-all` (`-pass` and `-skip` combined) we can get additional info, such as which tests were skipped and elapsed time of each passed test.

`tparse` comes with a `-dump` flag to replay everything that would have otherwise been printed. Enabling users to retrieve original `go test` output. Eliminating the need for `tee /dev/tty` between pipes.

The default order is:
- `go test` output (if adding `-dump` flag)
- passed/skipped table (if adding `-all`, `-skip` or `-pass` flag)
- failed tests and panics
- summary

Default order can be reversed with `-top` flag.

For narrow displays the `-smallscreen` flag may be useful, dividing a long test name and making it vertical heavy:

```
TestSubtests/an_awesome_but_long/subtest_for_the/win

TestSubtests
 /an_awesome_but_long
 /subtest_for_the
 /win
 ```

`tparse` aims to provide a simply alternative to one-liner bash functions.

---

P.S. `tparse` uses itself in travis [travis pipeline](https://travis-ci.com/mfridman/tparse/jobs/156185520):

<img src="https://www.dropbox.com/s/x9cva17f3ko82gb/travis-ci.png?raw=1" />