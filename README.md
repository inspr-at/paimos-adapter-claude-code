# paimos-adapter-claude-code

Reference PAIMOS harness adapter for Claude Code.

This repository owns the external `claude-code` adapter described by the
PAIMOS adapter protocol. The binary is a pure transform:

```sh
paimos-adapter-claude-code describe
paimos-adapter-claude-code validate --input -
paimos-adapter-claude-code render --input -
```

It consumes the PAIMOS canonical agent artifact JSON and writes Claude Code
command markdown for `.claude/commands/<slash>.md`. PAIMOS injects the managed
drift-detection header; the adapter never emits that header itself.

## Build

```sh
go test ./...
go build -o paimos-adapter-claude-code .
```

## Local discovery

PAIMOS discovers external adapters one directory below `$PAIMOS_ADAPTER_PATH`.
With this checkout at `~/Code/paimos-adapter-claude-code`:

```sh
cd ~/Code/paimos-adapter-claude-code
go build -o paimos-adapter-claude-code .

export PAIMOS_ADAPTER_PATH="$HOME/Code"
paimos skill list-adapters
paimos skill test-adapter claude-code
```

The checkout directory intentionally matches the external repo name while the
manifest name remains `claude-code`, so the executable next to the manifest is
`paimos-adapter-claude-code`.

## Conformance

This repo carries `expected_output.txt`, the optional PAIMOS conformance
snapshot for the representative fixture. CI builds this adapter, builds the
current PAIMOS CLI from `main`, and runs:

```sh
PAIMOS_ADAPTER_PATH="$(dirname "$PWD")" paimos skill test-adapter claude-code
```

Releases are versioned independently from PAIMOS releases. Bump
`paimos-adapter.json` on adapter output-format changes and tag this repository
with its own SemVer, for example `v1.0.1`.

## License

AGPL-3.0-only, matching PAIMOS.
