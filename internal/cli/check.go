package cli

import (
	"flag"
	"fmt"
	"os"

	"github.com/vdemeester/praetorian/internal/config"
	"github.com/vdemeester/praetorian/internal/engine"
)

// checkCmd is a read-only diagnostic tool with two modes:
//
//   - config validation (default): load config, list aliases
//   - command simulation (--alias + --command): evaluate a command against an
//     alias and report ALLOWED / DENIED
func checkCmd(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config file")
	alias := fs.String("alias", "", "alias to simulate against")
	command := fs.String("command", "", "command string to simulate (as SSH_ORIGINAL_COMMAND)")
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
