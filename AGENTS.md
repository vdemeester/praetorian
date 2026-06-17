# AGENTS.md

Guidance for AI agents and contributors working in this repository.

## What this is

Praetorian is an **SSH command restrictor**: a binary used as the target of a
`command="praetorian run <alias>"` directive in `authorized_keys`. It validates
`SSH_ORIGINAL_COMMAND` against a per-alias allow-list and executes it directly,
**without a shell**.

This is the 2.0 rewrite of the 2014–2016 implementation (preserved in git
history, tagged `0.1.2`).

## Core invariants — do not break these

- **No shell, ever.** Commands are tokenized with `google/shlex` and run via
  `syscall.Exec`. Never pass a command through `sh -c` or `exec.Command` with a
  shell. `$()`, backticks, `;`, `|` must stay inert bytes.
- **Default-deny.** Anything not explicitly allowed is denied. There are no
  top-level deny rules; denial exists only as narrowing constraints
  (`no_arg`) inside allow rules.
- **TOCTOU-free.** What is validated must be exactly what is exec'd — no
  re-tokenization or shell re-interpretation between check and exec.
- **Terse denials.** User-facing output on denial is `praetorian: denied` only.
  Details go to logs (`log/slog`), never to the client.

## Layout

| Path                   | Responsibility                                        |
| ---------------------- | ----------------------------------------------------- |
| `main.go`              | thin entrypoint → `internal/cli`                      |
| `internal/cli`         | subcommand dispatch (`run`, `check`, `version`)       |
| `internal/config`      | HCL/JSON config schema + loading                      |
| `internal/engine`      | tokenize + evaluate command against allow rules       |
| `internal/authkeys`    | parse/classify `authorized_keys` for `check`          |
| `version`              | build-time version vars (set via ldflags)             |

## Dependencies — keep minimal

Only `hashicorp/hcl/v2`, `google/shlex`, and the Go stdlib (`log/slog`,
`syscall`, `path`). Do not add CLI frameworks (no cobra/viper) — `os.Args`
switching is sufficient. Justify any new dependency in the PR.

## Workflow

- **TDD is required.** Write a failing test first (RED), minimal implementation
  (GREEN), then refactor. New/changed code is gated at 80% patch coverage by
  Codecov.
- **All work goes through PRs** off `main`. `main` is protected and requires the
  `build`, `lint`, and `test` checks.
- **Merge with rebase** (the repo is rebase-only, linear history).
- Use a **git worktree** per branch rather than switching in place.

## Commands

```sh
make build    # static binary into ./bin
make test     # go test -race -cover ./...
make lint     # golangci-lint (v2 config)
make check    # fmt + vet + lint + test
make snapshot # local goreleaser build, no publish
```

## Conventions

- Config is HCL but must remain JSON-parseable (Nix generates JSON). Don't use
  HCL features that have no JSON encoding.
- Keep `gosec`/`revive` clean; annotate the intentional `syscall.Exec` /
  file-open findings with a justified `//nolint` comment, don't disable linters
  globally.
- Releases are cut by pushing a `v*` tag (goreleaser).
