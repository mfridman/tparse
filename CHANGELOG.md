# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Add a `-follow-output` flag to allow writing go test output directly into a file. This will be
  useful (especially in CI jobs) for outputting overly verbose testing output into a file instead of
  the standard stream. (#133)

  | flag combination         | `go test` output destination |
  | ------------------------ | ---------------------------- |
  | No flags                 | Discard output               |
  | `-follow`                | Write to stdout              |
  | `-follow-output`         | Write to file                |
  | `-follow -follow-output` | Write to file                |

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

[Unreleased]: https://github.com/mfridman/tparse/compare/v0.15.0...HEAD
[v0.15.0]: https://github.com/mfridman/tparse/compare/v0.14.0...v0.15.0
[v0.14.0]: https://github.com/mfridman/tparse/compare/v0.13.3...v0.14.0
[v0.13.3]: https://github.com/mfridman/tparse/compare/v0.13.2...v0.13.3
[v0.13.2]: https://github.com/mfridman/tparse/compare/v0.13.1...v0.13.2
[v0.13.1]: https://github.com/mfridman/tparse/compare/v0.13.0...v0.13.1
[v0.13.0]: https://github.com/mfridman/tparse/releases/tag/v0.13.0
