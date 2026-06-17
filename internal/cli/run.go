package cli

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"syscall"

	"github.com/vdemeester/praetorian/internal/config"
	"github.com/vdemeester/praetorian/internal/engine"
)

// runCmd is the production gate, used as the `command=` target in
// authorized_keys. It reads SSH_ORIGINAL_COMMAND, validates it against the
// alias's allow rules, and execs the command directly or denies.
func runCmd(args []string) int {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	cfgPath := fs.String("config", "", "path to config file")
	logFormat := fs.String("log-format", "text", "log format: text or json")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "praetorian run: exactly one <alias> required")
		return 2
	}
	alias := rest[0]
	log := newLogger(*logFormat)

	path, err := resolveConfigPath(*cfgPath)
	if err != nil {
		log.Error("config error", "error", err)
		denied()
		return 1
	}
	cfg, err := config.Load(path)
	if err != nil {
		log.Error("config error", "config", path, "error", err)
		denied()
		return 1
	}
	log.Info("config loaded", "config", path, "aliases_count", len(cfg.Aliases))

	a := cfg.Lookup(alias)
	if a == nil {
		log.Error("alias not found", "alias", alias, "result", "DENIED", "reason", "alias not in config")
		denied()
		return 1
	}

	raw := os.Getenv("SSH_ORIGINAL_COMMAND")
	tokens, err := engine.Tokenize(raw)
	if err != nil {
		log.Warn("tokenize failed", "alias", alias, "result", "DENIED", "reason", err)
		denied()
		return 1
	}

	matched, err := engine.Evaluate(a, tokens)
	if err != nil {
		log.Warn("command denied", "alias", alias, "command", first(tokens), "args", argsOf(tokens), "result", "DENIED", "reason", err)
		denied()
		return 1
	}
	log.Info("command allowed", "alias", alias, "command", tokens[0], "args", argsOf(tokens), "matched_rule", matched.Command, "result", "ALLOWED")

	return execCommand(log, tokens)
}

// execCommand replaces the current process with the validated command.
func execCommand(log *slog.Logger, tokens []string) int {
	bin, err := exec.LookPath(tokens[0])
	if err != nil {
		log.Error("command not found", "command", tokens[0], "error", err)
		denied()
		return 1
	}
	// syscall.Exec replaces the process image; on success it does not return.
	// The command was validated against the allow-list above, which is the
	// entire purpose of praetorian.
	//nolint:gosec // G204: executing a validated, allow-listed command is intended
	if err := syscall.Exec(bin, tokens, os.Environ()); err != nil {
		log.Error("exec failed", "command", bin, "error", err)
		denied()
		return 1
	}
	return 0 // unreachable
}

// denied writes the terse, information-free denial message to stderr.
func denied() { fmt.Fprintln(os.Stderr, "praetorian: denied") }

func newLogger(format string) *slog.Logger {
	var h slog.Handler
	if format == "json" {
		h = slog.NewJSONHandler(os.Stderr, nil)
	} else {
		h = slog.NewTextHandler(os.Stderr, nil)
	}
	return slog.New(h)
}

func first(tokens []string) string {
	if len(tokens) == 0 {
		return ""
	}
	return tokens[0]
}

func argsOf(tokens []string) []string {
	if len(tokens) <= 1 {
		return nil
	}
	return tokens[1:]
}
