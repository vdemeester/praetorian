// Package cli implements praetorian's command-line dispatch. There are only a
// handful of subcommands, so os.Args switching is used rather than a CLI
// framework.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

// Main dispatches to a subcommand and returns a process exit code.
func Main(args []string) int {
	if len(args) == 0 {
		usage(os.Stderr)
		return 2
	}
	switch args[0] {
	case "run":
		return runCmd(args[1:])
	case "check":
		return checkCmd(args[1:])
	case "version":
		return versionCmd(args[1:])
	case "-h", "--help", "help":
		usage(os.Stdout)
		return 0
	default:
		fmt.Fprintf(os.Stderr, "praetorian: unknown subcommand %q\n", args[0])
		usage(os.Stderr)
		return 2
	}
}

func usage(w *os.File) {
	_, _ = fmt.Fprint(w, `praetorian — SSH command restrictor

Usage:
  praetorian run [--config PATH] <alias>     Validate SSH_ORIGINAL_COMMAND and exec
  praetorian check [--config PATH] [...]      Diagnostics: validate / simulate
  praetorian version                          Print version information

Config lookup order (first found wins, no merge):
  1. --config PATH
  2. ~/.config/praetorian/config.hcl, then config.json
  3. /etc/praetorian/config.hcl, then config.json
`)
}

// resolveConfigPath implements the documented lookup order. An explicit path
// always wins (and is returned even if missing, so the error is clear).
//
// At each location HCL (human-written) is preferred over JSON (Nix-generated),
// so a hand-edited config wins over a generated one in the same directory.
func resolveConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	candidates := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		dir := filepath.Join(home, ".config", "praetorian")
		candidates = append(candidates,
			filepath.Join(dir, "config.hcl"),
			filepath.Join(dir, "config.json"),
		)
	}
	candidates = append(candidates,
		"/etc/praetorian/config.hcl",
		"/etc/praetorian/config.json",
	)
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("no config found (looked in: %v)", candidates)
}
