# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Update dependencies to latest versions
- Include failed test names in table (#149)
- Colorize the status in `--progress` output (#150)

## [v0.18.0] - 2025-08-24

- Wrap panic messages at the terminal width (#142)
- Do not include packages with no coverage in the output (#144)

## [v0.17.0]

- Deprecate github.com/mfridman/buildversion, and use std lib `debug.ReadBuildInfo()` instead. In
  go1.24 this is handled automatically, from the [release notes](https://go.dev/doc/go1.24):

  > The go build command now sets the main module’s version in the compiled binary based on the
  > version control system tag and/or commit. A +dirty suffix will be appended if there are
  > uncommitted changes. Use the -buildvcs=false flag to omit version control information from the
  > binary.

- Handle changes in go1.24 related to build output. `tparse` will pipe the build output to stderr

  > Furthermore, `go test -json` now reports build output and failures in JSON, interleaved with
  > test result JSON. These are distinguished by new Action types, but if they cause problems in a
  > test integration system, you can revert to the text build output with GODEBUG setting
  > gotestjsonbuildtext=1.

## [v0.16.0]

- Add a `-follow-output` flag to allow writing go test output directly into a file. This will be
  useful (especially in CI jobs) for outputting overly verbose testing output into a file instead of
  the standard stream. (#134)

  | flag combination         | `go test` output destination |
  | ------------------------ | ---------------------------- |
  | No flags                 | Discard output               |
  | `-follow`                | Write to stdout              |
  | `-follow-output`         | Write to file                |
  | `-follow -follow-output` | Write to file                |

- Use [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) for table rendering.
  - This will allow for more control over the output and potentially more features in the future.
    (#136)
  - Minor changes to the output format are expected, but the overall content should remain the same.
    If you have any feedback, please let me know.

## [v0.15.0]

- Add `-trimpath` flag, which removes the path prefix from package names in the output, simplifying
  their display. See #128 for examples.
  - There's a special case for `-trimpath=auto` which will automatically determine the prefix based
    on the longest common prefix of all package paths.

## [v0.14.0]

- Modify `--follow` behavior by minimizing noisy output. (#122)

> [!TIP]
>
> If you want the existing behavior, I added a `--follow-verbose` flag. But please do let me know if
> this affected you, as I plan to remove this before cutting a `v1.0.0`. Thank you!

## [v0.13.3]

- General housekeeping and dependency updates.

## [v0.13.2]

- Add partial support for `-compare`. A feature that displays the coverage difference against a
  previous run. See description for more details
  https://github.com/mfridman/tparse/pull/101#issue-1857786730 and the initial issue #92.
- Fix unstable common package prefix logic #104

## [v0.13.1] - 2023-08-04

- Fix failing GoReleaser GitHub action (release notes location).

Summary from [v0.13.0](https://github.com/mfridman/tparse/releases/tag/v0.13.0)

- Start a [CHANGELOG.md](https://github.com/mfridman/tparse/blob/main/CHANGELOG.md) for user-facing
  change.
- Add [GoReleaser](https://goreleaser.com/) to automate the release process. Pre-built binaries are
  available for each release, currently Linux and macOS. If there is demand, can also add Windows.

## [v0.13.0] - 2023-08-04

- Start a [CHANGELOG.md](https://github.com/mfridman/tparse/blob/main/CHANGELOG.md) for user-facing
  change.
- Add [GoReleaser](https://goreleaser.com/) to automate the release process. Pre-built binaries are
  available for each release, currently Linux and macOS. If there is demand, can also add Windows.

[Unreleased]: https://github.com/mfridman/tparse/compare/v0.18.0...HEAD
[v0.18.0]: https://github.com/mfridman/tparse/compare/v0.17.0...v0.18.0
[v0.17.0]: https://github.com/mfridman/tparse/compare/v0.16.0...v0.17.0
[v0.16.0]: https://github.com/mfridman/tparse/compare/v0.15.0...v0.16.0
[v0.15.0]: https://github.com/mfridman/tparse/compare/v0.14.0...v0.15.0
[v0.14.0]: https://github.com/mfridman/tparse/compare/v0.13.3...v0.14.0
[v0.13.3]: https://github.com/mfridman/tparse/compare/v0.13.2...v0.13.3
[v0.13.2]: https://github.com/mfridman/tparse/compare/v0.13.1...v0.13.2
[v0.13.1]: https://github.com/mfridman/tparse/compare/v0.13.0...v0.13.1
[v0.13.0]: https://github.com/mfridman/tparse/releases/tag/v0.13.0
