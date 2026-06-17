package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/vdemeester/praetorian/internal/authkeys"
	"github.com/vdemeester/praetorian/internal/config"
	"github.com/vdemeester/praetorian/internal/engine"
)

// checkCmd is a read-only diagnostic tool with three modes:
//
//   - config validation (default): load config, list aliases
//   - command simulation (--alias + --command): evaluate a command against an
//     alias and report ALLOWED / DENIED
//   - authorized_keys cross-check (--authorized-keys): classify each key line
//     against the config
func checkCmd(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config file")
	alias := fs.String("alias", "", "alias to simulate against")
	command := fs.String("command", "", "command string to simulate (as SSH_ORIGINAL_COMMAND)")
	authKeys := fs.String("authorized-keys", "", "path to an authorized_keys file to cross-check")
	strict := fs.Bool("strict", false, "treat informational notes as warnings (non-zero exit on any warning)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	path, err := resolveConfigPath(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		return 1
	}
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		return 1
	}

	if *alias != "" || *command != "" {
		return simulate(cfg, *alias, *command)
	}
	if *authKeys != "" {
		return crossCheck(cfg, *authKeys, *strict)
	}

	fmt.Printf("✓ Config loaded: %s\n", path)
	names := make([]string, 0, len(cfg.Aliases))
	for _, a := range cfg.Aliases {
		names = append(names, a.Name)
	}
	fmt.Printf("✓ %d aliases defined: %v\n", len(cfg.Aliases), names)
	return 0
}

func simulate(cfg *config.Config, alias, command string) int {
	if alias == "" || command == "" {
		fmt.Fprintln(os.Stderr, "✗ both --alias and --command are required for simulation")
		return 2
	}
	a := cfg.Lookup(alias)
	if a == nil {
		fmt.Printf("✗ Alias: %s (not in config)\n→ DENIED\n", alias)
		return 1
	}
	tokens, err := engine.Tokenize(command)
	if err != nil {
		fmt.Printf("✗ Command: %v\n→ DENIED\n", err)
		return 1
	}
	matched, err := engine.Evaluate(a, tokens)
	if err != nil {
		fmt.Printf("✗ Alias: %s\n✗ Command: %s\n✗ %v\n→ DENIED\n", alias, first(tokens), err)
		return 1
	}
	fmt.Printf("✓ Alias: %s\n✓ Command: %s\n✓ Args: %v\n✓ Matched: allow %q\n→ ALLOWED\n", alias, tokens[0], argsOf(tokens), matched.Command)
	return 0
}

// crossCheck classifies every key in an authorized_keys file against the
// config. Returns non-zero if any key would be denied (missing alias), or on
// any informational note when --strict is set.
func crossCheck(cfg *config.Config, path string, strict bool) int {
	f, err := os.Open(path) //nolint:gosec // path is an operator-supplied diagnostic input
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		return 1
	}

	entries, err := authkeys.Analyze(f, cfg)
	_ = f.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "✗ %v\n", err)
		return 1
	}

	warnings, infos := 0, 0
	for _, e := range entries {
		switch e.Kind {
		case authkeys.PraetorianFound:
			fmt.Printf("✓ %-16s → config found, %d allow rules\n", e.Alias, e.AllowCount)
		case authkeys.PraetorianMissing:
			warnings++
			fmt.Printf("⚠ %-16s → alias not in config; this key will be DENIED for all commands\n", e.Alias)
		case authkeys.Unrestricted:
			infos++
			fmt.Printf("ℹ key without command= → unrestricted access\n")
		case authkeys.OtherCommand:
			infos++
			fmt.Printf("ℹ key with command=%q → not praetorian\n", e.Command)
		}
	}

	if warnings > 0 {
		return 1
	}
	if strict && infos > 0 {
		return 1
	}
	return 0
}
