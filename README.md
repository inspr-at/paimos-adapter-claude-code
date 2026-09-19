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

## Contributing

### Developer Certificate of Origin

New contributions use the unmodified [Developer Certificate of Origin 1.1](DCO).
A `Signed-off-by: Name <email>` trailer records that you have the right to submit
the contribution under the project licence. It is not a cryptographic signature
or a guarantee of correctness. Use the same identity as the commit author;
a GitHub-associated noreply address is fine. Sign-offs remain in public history.
This applies to maintainers and outside contributors alike, from adoption onward;
existing history is not rewritten. After reading the DCO, create each new commit
with `git commit -s` using your own name and GitHub-associated email.

Fork the repository, create a branch from the current upstream `main`, implement
and test your change, then push to your fork and open a pull request to `main`.
Describe the change, its purpose, tests, and any limitations. Contributors need
no write access to this repository. The maintainer reviews agent findings and
decides whether to merge; passing checks never grants an agent merge authority.

The required `dco` check validates every commit introduced by a PR, including
merge commits on the contributor branch. An empty or incomplete range fails.
Local checks require Python 3 and full Git history; deepen a shallow checkout
with `git fetch --unshallow` first. When merging upstream updates into your
branch, use `git merge --signoff upstream/main` after fetching upstream.
It reads real Git trailers, so a sign-off quoted in prose does not count.
Missing sign-offs must be supplied by the contributor, not invented by a reviewer
or agent. Do not rewrite shared history to repair them without explicit agreement.

Bots are not exempt. Dependabot's native `Signed-off-by` service address is
accepted for its exact GitHub author identity; other bots use their own matching
author/sign-off identity. This checks declarations, not account authenticity.
For agent-assisted work, the human contributor must understand and authorize
their DCO declaration; the agent must not invent identities or sign for others.

GitHub web commits require sign-off. For squash merges, retain the original
commit messages and move their existing sign-off declarations into the final
trailer block; an indented or quoted sign-off is not a trailer. Check that the
final author still has a matching declaration. Never invent a contributor's
sign-off. Use a regular merge when combining authors would obscure provenance.
Release and deployment remain maintainer-controlled. Existing review and CI
requirements still apply; DCO introduces no second-maintainer requirement.

## License

AGPL-3.0-only, matching PAIMOS.
