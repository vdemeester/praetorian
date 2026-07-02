package cli

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// fakeSyslog records the last message and the severity method used.
type fakeSyslog struct {
	level string
	msg   string
}

func (f *fakeSyslog) Debug(m string) error   { f.level, f.msg = "debug", m; return nil }
func (f *fakeSyslog) Info(m string) error    { f.level, f.msg = "info", m; return nil }
func (f *fakeSyslog) Warning(m string) error { f.level, f.msg = "warning", m; return nil }
func (f *fakeSyslog) Err(m string) error     { f.level, f.msg = "err", m; return nil }

// The syslog handler must map slog levels to the matching syslog severity so
// WARN/ERROR audit records stay filterable in the system logger.
func TestSyslogHandler_MapsLevels(t *testing.T) {
	cases := []struct {
		log   func(*slog.Logger)
		level string
	}{
		{func(l *slog.Logger) { l.Info("config loaded") }, "info"},
		{func(l *slog.Logger) { l.Warn("command denied") }, "warning"},
		{func(l *slog.Logger) { l.Error("exec failed") }, "err"},
	}
	for _, c := range cases {
		f := &fakeSyslog{}
		log := slog.New(newSyslogHandler(f, "text", slog.LevelInfo))
		c.log(log)
		if f.level != c.level {
			t.Errorf("expected syslog severity %q, got %q (msg=%q)", c.level, f.level, f.msg)
		}
	}
}

// Records must not carry slog's own timestamp (syslog/journald adds one) and
// must not end with a trailing newline (syslog is message-based).
func TestSyslogHandler_NoTimestampNoTrailingNewline(t *testing.T) {
	f := &fakeSyslog{}
	log := slog.New(newSyslogHandler(f, "text", slog.LevelInfo))
	log.Info("config loaded", "aliases_count", 3)

	if strings.Contains(f.msg, "time=") {
		t.Errorf("expected no time= attribute, got %q", f.msg)
	}
	if strings.HasSuffix(f.msg, "\n") {
		t.Errorf("expected no trailing newline, got %q", f.msg)
	}
	if !strings.Contains(f.msg, "config loaded") || !strings.Contains(f.msg, "aliases_count=3") {
		t.Errorf("expected record content, got %q", f.msg)
	}
}

// JSON format is honoured by the syslog handler.
func TestSyslogHandler_JSON(t *testing.T) {
	f := &fakeSyslog{}
	log := slog.New(newSyslogHandler(f, "json", slog.LevelInfo))
	log.Warn("command denied", "alias", "aomi-git")

	if !strings.HasPrefix(f.msg, "{") || !strings.Contains(f.msg, `"command denied"`) {
		t.Errorf("expected JSON record, got %q", f.msg)
	}
}

// Enabled must delegate to the inner handler (INFO is on by default).
func TestSyslogHandler_Enabled(t *testing.T) {
	h := newSyslogHandler(&fakeSyslog{}, "text", slog.LevelInfo)
	if !h.Enabled(context.Background(), slog.LevelInfo) {
		t.Fatal("expected INFO to be enabled")
	}
}

// WithAttrs/WithGroup must preserve the syslog dispatch and attach attributes.
func TestSyslogHandler_WithAttrsGroup(t *testing.T) {
	f := &fakeSyslog{}
	log := slog.New(newSyslogHandler(f, "text", slog.LevelInfo)).With("alias", "aomi-git").WithGroup("g")
	log.Warn("command denied", "reason", "no match")

	if f.level != "warning" {
		t.Errorf("expected warning severity, got %q", f.level)
	}
	if !strings.Contains(f.msg, "alias=aomi-git") || !strings.Contains(f.msg, "g.reason=") {
		t.Errorf("expected attrs and group in output, got %q", f.msg)
	}
}

// Default audit handler must be the syslog handler when the socket is
// reachable — never a stderr handler that would leak to the client.
func TestAuditHandler_DefaultIsSyslog(t *testing.T) {
	t.Setenv("PRAETORIAN_LOG_STDERR", "")
	orig := syslogDial
	syslogDial = func() (syslogWriter, error) { return &fakeSyslog{}, nil }
	t.Cleanup(func() { syslogDial = orig })

	if _, ok := auditHandler("text", slog.LevelInfo).(*syslogHandler); !ok {
		t.Fatal("expected *syslogHandler when syslog is reachable")
	}
}

// When the socket is unavailable, records are discarded (not sent to stderr).
func TestAuditHandler_FallbackOnDialError(t *testing.T) {
	t.Setenv("PRAETORIAN_LOG_STDERR", "")
	orig := syslogDial
	syslogDial = func() (syslogWriter, error) { return nil, errors.New("no syslog") }
	t.Cleanup(func() { syslogDial = orig })

	if _, ok := auditHandler("text", slog.LevelInfo).(*syslogHandler); ok {
		t.Fatal("expected non-syslog (discard) handler on dial error")
	}
}

// The debugging escape hatch routes to a plain stderr handler.
func TestAuditHandler_StderrOptIn(t *testing.T) {
	t.Setenv("PRAETORIAN_LOG_STDERR", "1")
	if _, ok := auditHandler("text", slog.LevelInfo).(*syslogHandler); ok {
		t.Fatal("expected a plain (stderr) handler when PRAETORIAN_LOG_STDERR set")
	}
}

// Debug records map to the syslog debug severity.
func TestSyslogHandler_DebugLevel(t *testing.T) {
	f := &fakeSyslog{}
	h := newSyslogHandler(f, "text", slog.LevelInfo)
	r := slog.NewRecord(time.Time{}, slog.LevelDebug, "dbg", 0)
	if err := h.Handle(context.Background(), r); err != nil {
		t.Fatal(err)
	}
	if f.level != "debug" {
		t.Errorf("expected debug severity, got %q", f.level)
	}
}

// textHandler honours the json format selector for the stderr/discard paths.
func TestTextHandler_JSON(t *testing.T) {
	var buf bytes.Buffer
	log := slog.New(textHandler("json", slog.LevelInfo, &buf))
	log.Info("hello")
	if !strings.HasPrefix(strings.TrimSpace(buf.String()), "{") {
		t.Errorf("expected JSON output, got %q", buf.String())
	}
}

// parseLevel resolves flag, then env, then defaults to Info; unknown values
// also fall back to Info.
func TestParseLevel(t *testing.T) {
	t.Setenv("PRAETORIAN_LOG_LEVEL", "")
	cases := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"info":    slog.LevelInfo,
		"warn":    slog.LevelWarn,
		"warning": slog.LevelWarn,
		"ERROR":   slog.LevelError,
		"":        slog.LevelInfo,
		"bogus":   slog.LevelInfo,
	}
	for in, want := range cases {
		if got := parseLevel(in); got != want {
			t.Errorf("parseLevel(%q) = %v, want %v", in, got, want)
		}
	}
}

// An empty flag falls back to the PRAETORIAN_LOG_LEVEL env var.
func TestParseLevel_EnvFallback(t *testing.T) {
	t.Setenv("PRAETORIAN_LOG_LEVEL", "warn")
	if got := parseLevel(""); got != slog.LevelWarn {
		t.Errorf("expected env fallback to warn, got %v", got)
	}
	// An explicit flag wins over the env var.
	if got := parseLevel("debug"); got != slog.LevelDebug {
		t.Errorf("expected flag to win, got %v", got)
	}
}

// The configured level must actually suppress lower-severity records: at Warn,
// an Info record is dropped and never reaches the syslog sink.
func TestSyslogHandler_LevelFilters(t *testing.T) {
	f := &fakeSyslog{}
	log := slog.New(newSyslogHandler(f, "text", slog.LevelWarn))
	log.Info("config loaded")
	if f.level != "" {
		t.Errorf("expected Info to be suppressed at Warn level, got %q/%q", f.level, f.msg)
	}
	log.Warn("command denied")
	if f.level != "warning" {
		t.Errorf("expected Warn to pass, got %q", f.level)
	}
}

// newLogger must return a usable logger.
func TestNewLogger_Usable(t *testing.T) {
	t.Setenv("PRAETORIAN_LOG_STDERR", "")
	orig := syslogDial
	syslogDial = func() (syslogWriter, error) { return &fakeSyslog{}, nil }
	t.Cleanup(func() { syslogDial = orig })

	if newLogger("text", slog.LevelInfo) == nil {
		t.Fatal("expected a logger")
	}
}
