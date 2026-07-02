package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"log/syslog"
	"os"
	"os/exec"
	"strings"
	"sync"
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
	logLevel := fs.String("log-level", "", "log level: debug, info, warn, error (env: PRAETORIAN_LOG_LEVEL, default info)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(os.Stderr, "praetorian run: exactly one <alias> required")
		return 2
	}
	alias := rest[0]
	log := newLogger(*logFormat, parseLevel(*logLevel))

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

// newLogger builds the audit logger. The handler is chosen by auditHandler so
// that operational INFO/WARN records are routed to a server-side sink and do
// not leak onto the client's SSH stderr channel.
func newLogger(format string, level slog.Level) *slog.Logger {
	return slog.New(auditHandler(format, level))
}

// parseLevel resolves the effective log level from the --log-level flag,
// falling back to the PRAETORIAN_LOG_LEVEL env var (useful because the
// authorized_keys command line is fixed), then to Info. Unknown values also
// fall back to Info.
func parseLevel(flagValue string) slog.Level {
	s := flagValue
	if s == "" {
		s = os.Getenv("PRAETORIAN_LOG_LEVEL")
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// auditHandler decides where structured audit records go.
//
// Because sshd runs `praetorian run` as the remote command, os.Stderr is the
// client's terminal. Audit records must therefore go to a server-side sink:
//   - default: the local syslog/journald socket (facility AUTHPRIV, since
//     these are authorization decisions), with slog levels mapped to syslog
//     severities;
//   - if the socket is unavailable, records are discarded rather than leaked
//     to the client;
//   - PRAETORIAN_LOG_STDERR (any non-empty value) is a debugging escape hatch
//     that routes records to stderr instead.
//
// The single user-facing `praetorian: denied` message (see denied) always
// goes to os.Stderr regardless of this destination.
func auditHandler(format string, level slog.Level) slog.Handler {
	if os.Getenv("PRAETORIAN_LOG_STDERR") != "" {
		return textHandler(format, level, os.Stderr)
	}
	w, err := syslogDial()
	if err != nil {
		return textHandler(format, level, io.Discard)
	}
	return newSyslogHandler(w, format, level)
}

func textHandler(format string, level slog.Level, dst io.Writer) slog.Handler {
	opts := &slog.HandlerOptions{Level: level}
	if format == "json" {
		return slog.NewJSONHandler(dst, opts)
	}
	return slog.NewTextHandler(dst, opts)
}

// syslogWriter is the subset of *syslog.Writer the handler needs; kept as an
// interface so tests can supply a fake.
type syslogWriter interface {
	Debug(m string) error
	Info(m string) error
	Warning(m string) error
	Err(m string) error
}

// syslogDial opens the audit sink. The base priority is facility-only
// (AUTHPRIV); the per-record severity is supplied by the handler via the
// Debug/Info/Warning/Err methods. It is a package variable so tests can
// substitute the system-log connection deterministically.
var syslogDial = func() (syslogWriter, error) {
	return syslog.New(syslog.LOG_AUTHPRIV, "praetorian")
}

// syslogHandler adapts slog to a syslog.Writer: it formats records with a
// standard slog handler (minus the timestamp, which syslog/journald supplies)
// and dispatches each formatted line to the syslog severity matching the slog
// level, so WARN/ERROR are filterable in the system logger.
type syslogHandler struct {
	w     syslogWriter
	mu    *sync.Mutex
	buf   *bytes.Buffer
	inner slog.Handler
}

func newSyslogHandler(w syslogWriter, format string, level slog.Level) *syslogHandler {
	buf := &bytes.Buffer{}
	opts := &slog.HandlerOptions{Level: level, ReplaceAttr: dropTime}
	var inner slog.Handler
	if format == "json" {
		inner = slog.NewJSONHandler(buf, opts)
	} else {
		inner = slog.NewTextHandler(buf, opts)
	}
	return &syslogHandler{w: w, mu: &sync.Mutex{}, buf: buf, inner: inner}
}

// dropTime removes the top-level time attribute; syslog/journald timestamps
// each record itself, so keeping slog's would duplicate it.
func dropTime(groups []string, a slog.Attr) slog.Attr {
	if len(groups) == 0 && a.Key == slog.TimeKey {
		return slog.Attr{}
	}
	return a
}

func (h *syslogHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.inner.Enabled(ctx, l)
}

func (h *syslogHandler) Handle(ctx context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.buf.Reset()
	if err := h.inner.Handle(ctx, r); err != nil {
		return err
	}
	// syslog records are message-based; strip the handler's single trailing
	// newline to avoid awkward multi-line entries in journalctl.
	msg := strings.TrimSuffix(h.buf.String(), "\n")
	switch {
	case r.Level >= slog.LevelError:
		return h.w.Err(msg)
	case r.Level >= slog.LevelWarn:
		return h.w.Warning(msg)
	case r.Level >= slog.LevelInfo:
		return h.w.Info(msg)
	default:
		return h.w.Debug(msg)
	}
}

func (h *syslogHandler) WithAttrs(as []slog.Attr) slog.Handler {
	return &syslogHandler{w: h.w, mu: h.mu, buf: h.buf, inner: h.inner.WithAttrs(as)}
}

func (h *syslogHandler) WithGroup(name string) slog.Handler {
	return &syslogHandler{w: h.w, mu: h.mu, buf: h.buf, inner: h.inner.WithGroup(name)}
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
