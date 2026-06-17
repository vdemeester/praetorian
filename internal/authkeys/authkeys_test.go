package authkeys

import (
	"strings"
	"testing"

	"github.com/vdemeester/praetorian/internal/config"
)

func cfg(aliases ...string) *config.Config {
	c := &config.Config{}
	for _, a := range aliases {
		c.Aliases = append(c.Aliases, config.Alias{Name: a})
	}
	return c
}

func TestAnalyze_MissingAlias(t *testing.T) {
	line := `command="praetorian run okinawa-tpm" ssh-ed25519 AAAAC3Nz vincent@kyushu`
	entries, err := Analyze(strings.NewReader(line), cfg())
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1", len(entries))
	}
	e := entries[0]
	if e.Kind != PraetorianMissing {
		t.Errorf("Kind = %v, want PraetorianMissing", e.Kind)
	}
	if e.Alias != "okinawa-tpm" {
		t.Errorf("Alias = %q, want okinawa-tpm", e.Alias)
	}
}

func TestAnalyze_FoundAlias(t *testing.T) {
	c := &config.Config{Aliases: []config.Alias{{
		Name:  "okinawa-tpm",
		Allow: []config.Allow{{Command: "borg serve"}, {Command: "git-upload-pack"}},
	}}}
	line := `command="praetorian run okinawa-tpm",no-pty ssh-ed25519 AAAAC3Nz vincent@kyushu`
	entries, err := Analyze(strings.NewReader(line), c)
	if err != nil {
		t.Fatalf("Analyze error: %v", err)
	}
	e := entries[0]
	if e.Kind != PraetorianFound || e.Alias != "okinawa-tpm" || e.AllowCount != 2 {
		t.Errorf("got %+v, want Found okinawa-tpm AllowCount=2", e)
	}
}

func TestAnalyze_ConfigFlagSkipped(t *testing.T) {
	line := `command="praetorian run --config /etc/p.hcl aomi-tpm" ssh-ed25519 AAAA x@y`
	entries, _ := Analyze(strings.NewReader(line), cfg("aomi-tpm"))
	if entries[0].Kind != PraetorianFound || entries[0].Alias != "aomi-tpm" {
		t.Errorf("got %+v, want Found aomi-tpm (flag skipped)", entries[0])
	}
}

func TestAnalyze_Unrestricted(t *testing.T) {
	line := `sk-ssh-ed25519@openssh.com AAAAGnNr vincent@yubikey`
	entries, _ := Analyze(strings.NewReader(line), cfg())
	if entries[0].Kind != Unrestricted {
		t.Errorf("Kind = %v, want Unrestricted", entries[0].Kind)
	}
}

func TestAnalyze_OtherCommand(t *testing.T) {
	line := `command="/usr/local/bin/deploy.sh" ssh-ed25519 AAAA deploy@ci`
	entries, _ := Analyze(strings.NewReader(line), cfg())
	if entries[0].Kind != OtherCommand || entries[0].Command != "/usr/local/bin/deploy.sh" {
		t.Errorf("got %+v, want OtherCommand deploy.sh", entries[0])
	}
}

func TestAnalyze_SkipsBlankAndComments(t *testing.T) {
	in := "# a comment\n\ncommand=\"praetorian run x\" ssh-ed25519 AAAA c\n"
	entries, _ := Analyze(strings.NewReader(in), cfg("x"))
	if len(entries) != 1 {
		t.Fatalf("got %d entries, want 1 (blanks/comments skipped)", len(entries))
	}
}
