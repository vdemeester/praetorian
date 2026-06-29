# Praetorian

[![CI](https://github.com/vdemeester/praetorian/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/vdemeester/praetorian/actions/workflows/ci.yaml)
[![codecov](https://codecov.io/gh/vdemeester/praetorian/branch/main/graph/badge.svg)](https://codecov.io/gh/vdemeester/praetorian)
[![Go Report Card](https://goreportcard.com/badge/github.com/vdemeester/praetorian)](https://goreportcard.com/report/github.com/vdemeester/praetorian)
[![Go Reference](https://pkg.go.dev/badge/github.com/vdemeester/praetorian.svg)](https://pkg.go.dev/github.com/vdemeester/praetorian)
[![License: GPL-3.0](https://img.shields.io/badge/License-GPLv3-blue.svg)](LICENSE)

<img src="imgs/praetorian.png" alt="Praetorian logo" title="The man himself" align="right" />

Praetorian is an **SSH command restrictor**. It is used as the target of a
`command="..."` directive in `authorized_keys` and validates the command a
client tries to run against a per-alias allow-list before executing it
**directly, without a shell**.

> **2.0 rewrite.** This is a ground-up rewrite of the original (2014–2016)
> praetorian. The legacy implementation is preserved in git history and tagged
> `0.1.2`.

## Security model

- **Default-deny** — anything not explicitly allowed is denied.
- **No shell, ever** — the command is tokenized with a shell-like lexer
  ([`google/shlex`](https://github.com/google/shlex)) and run via
  `syscall.Exec`. `$()`, backticks, `;`, `|` are inert bytes.
- **TOCTOU-free** — what praetorian validates is exactly what it execs; there is
  no shell re-interpretation gap.
- **Allow-only with narrowing** — there are no top-level deny rules; denial only
  exists as narrowing constraints inside allow rules.

## Usage

In `authorized_keys`:

```
command="praetorian run okinawa-tpm",no-pty,no-port-forwarding ssh-ed25519 AAAA... user@host
```

`sshd` sets `SSH_ORIGINAL_COMMAND`; praetorian validates it against the
`okinawa-tpm` alias and execs or denies.

### Configuration

Written in HCL (and equally parseable as JSON — Nix can generate the JSON form):

```hcl
alias "okinawa-tpm" {
  allow "borg serve" {}              # required token prefix, any trailing args

  allow "git-upload-pack" {
    arg { pos = 1, glob = "/srv/git/*" }
    num_args = 1
  }

  allow "rsync" {
    any_arg = "/srv/backup/*"          # at least one arg must match
    no_arg  = "/srv/backup/.secret/*"  # narrowing: no arg may match
  }
}
```

Config lookup order (first found wins, **no merge**):

1. `--config PATH`
2. `~/.config/praetorian/config.hcl`, then `config.json`
3. `/etc/praetorian/config.hcl`, then `config.json`

At each location the human-written HCL is preferred over a Nix-generated JSON
config in the same directory.

#### Constraints

| Constraint            | Syntax                          | Meaning                                            |
| --------------------- | ------------------------------- | -------------------------------------------------- |
| none                  | `allow "cmd" {}`                | command/prefix match only, any args                |
| positional arg        | `arg { pos = N, glob = "..." }` | arg at 1-based position N matches glob (`-1`=last) |
| any arg               | `any_arg = "..."`               | at least one arg matches                           |
| no arg (narrowing)    | `no_arg = "..."`                | no arg may match                                   |
| arg count             | `num_args = N`                  | exactly N args                                      |

The `allow` label is shlex-split into a required token prefix: the first token
is the executable, remaining tokens are required leading arguments; constraints
apply to the arguments that follow.

## Subcommands

- `praetorian run <alias>` — production gate (use in `authorized_keys`).
- `praetorian check` — diagnostics: validate config, simulate a command with
  `--alias` + `--command`, or cross-check an `authorized_keys` file with
  `--authorized-keys` (add `--strict` to fail on informational notes too).
- `praetorian version` — version info.

```console
$ praetorian check --config examples/config.hcl --alias okinawa-tpm --command "rm -rf /"
✗ Alias: okinawa-tpm
✗ Command: rm
✗ denied: prefix mismatch
→ DENIED
```

## Development

```sh
make build    # build the binary into ./bin
make test     # run unit tests
make lint     # golangci-lint
make check    # fmt + vet + lint + test
make snapshot # build a local goreleaser snapshot (no publish)
```

## Releasing

Releases are cut by [goreleaser](https://goreleaser.com) when a `v*` tag is
pushed:

```sh
git tag -a v2.0.0 -m "praetorian 2.0.0"
git push origin v2.0.0
```

The `release` workflow builds static `linux`/`darwin` `amd64`/`arm64` binaries,
produces checksums, and publishes a GitHub release. Validate config locally
with `make release-check`.

## License

See [LICENSE](LICENSE).
