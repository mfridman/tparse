# tparse  [![Actions](https://github.com/mfridman/tparse/workflows/CI/badge.svg)](https://github.com/mfridman/tparse) [![Coverage](http://gocover.io/_badge/github.com/mfridman/tparse/parse)](http://gocover.io/github.com/mfridman/tparse/parse)

A command line tool for analyzing and summarizing `go test` output.

**Don't forget to run `go test` with the `-json` flag.**

Pass            |  Fail
:-------------------------:|:-------------------------:
<img src="https://res.cloudinary.com/mfridman/image/upload/q_auto/v1600038958/projects/tparse/passed_rlnd0i.png" />  |  <img src="https://res.cloudinary.com/mfridman/image/upload/q_auto/v1600038958/projects/tparse/failed_zdka7h.png" />

By default, `tparse` will always return test failures and panics, if any, followed by a package-level summary table.

To get additional info on passed tests run `tparse` with `-pass` flag. Tests are grouped by package and sorted by elapsed time in descending order (longest to shortest).

### [But why?!](#but-why) for more info.

## Installation

    go get github.com/mfridman/tparse

Or download the latest pre-built binary [here](https://github.com/mfridman/tparse/releases/latest).

## Usage

Once `tparse` is installed there are 2 ways to use it:

1. Run `go test` as normal, but add `-json` flag and pipe output to `tparse`.

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

`tparse` attempts to do just that; return failed tests and panics, if any, followed by a single package-level summary.

But, let's take it a bit further. With `-all` (`-pass` and `-skip` combined) can get additional info, such as which tests were skipped and elapsed time of each passed test.

`tparse` comes with a `-dump` flag to replay everything that would have otherwise been printed. Enabling users to retrieve original `go test` output. Eliminating the need for `tee /dev/tty` between pipes.

The default print order is:
- `go test` output (if adding `-dump` flag)
- passed/skipped table (if adding `-all`, `-skip` or `-pass` flag)
- failed tests and panics
- summary

The default print order can be reversed with `-top` flag.

For narrow displays the `-smallscreen` flag may be useful, dividing a long test name and making it vertical heavy:

```
TestSubtests/an_awesome_but_long/subtest_for_the/win

TestSubtests
 /an_awesome_but_long
 /subtest_for_the
 /win
 ```

`tparse` aims to be a simply alternative to one-liner bash functions.

---

P.S. `tparse` uses itself in [GitHub actions](https://github.com/mfridman/tparse/commit/eb87ddcaa52ed83692b01f6e30f3bd98aee036a3/checks?check_suite_id=345829033#step:5:11):

<img src="https://res.cloudinary.com/mfridman/image/upload/v1575645347/projects/tparse/Screen_Shot_2019-12-06_at_10.15.22_AM_itviiy.png" />