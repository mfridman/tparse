---
description: Getting started with tparse
---

# Getting started

If you have Go installed, try this out:

```bash
go install github.com/mfridman/tparse@latest
go test fmt strings bytes sort -cover -json -count=1 | tparse
```

The first command will install the tparse tool to your `$GOBIN` location (usually `$HOME/go/bin`).

The second command will run tests for familiar Go packages and pipe the output to tparse. 

!!! info ""

    **Make sure to run `go test` with the `-json` flag**

This is where the magic happens 🪄. tparse will parse the JSON output and return a summarized table of all the packages, time elapsed, number of tests and their status. Example:

```md
┌───────────────────────────────────────────────────────────┐
│  STATUS │ ELAPSED │ PACKAGE │ COVER │ PASS │ FAIL │ SKIP  │
│─────────┼─────────┼─────────┼───────┼──────┼──────┼───────│
│  PASS   │  1.81s  │ bytes   │ 95.6% │ 135  │  0   │  0    │
│  PASS   │  0.90s  │ fmt     │ 95.2% │  75  │  0   │  1    │
│  PASS   │  1.90s  │ sort    │ 60.8% │  38  │  0   │  1    │
│  PASS   │  1.41s  │ strings │ 98.1% │ 115  │  0   │  0    │
└───────────────────────────────────────────────────────────┘
```
