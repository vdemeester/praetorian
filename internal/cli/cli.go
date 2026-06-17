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
  2. ~/.config/praetorian/config.hcl
  3. /etc/praetorian/config.hcl
`)
}

// resolveConfigPath implements the documented lookup order. An explicit path
// always wins (and is returned even if missing, so the error is clear).
func resolveConfigPath(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	candidates := []string{}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "praetorian", "config.hcl"))
	}
	candidates = append(candidates, "/etc/praetorian/config.hcl")
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("no config found (looked in: %v)", candidates)
}
