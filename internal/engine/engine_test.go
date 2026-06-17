package engine

import (
	"errors"
	"testing"

	"github.com/vdemeester/praetorian/internal/config"
)

func ptr[T any](v T) *T { return &v }

func TestTokenize(t *testing.T) {
	tests := []struct {
		raw  string
		want []string
	}{
		{`git-upload-pack '/srv/git/repo.git'`, []string{"git-upload-pack", "/srv/git/repo.git"}},
		{`rsync --server -vlogDtpre.iLsfxCIvu . /srv/backup/`, []string{"rsync", "--server", "-vlogDtpre.iLsfxCIvu", ".", "/srv/backup/"}},
		{`borg serve`, []string{"borg", "serve"}},
		// Injection attempts are inert: shlex returns literal tokens, no shell.
		{`ls; rm -rf /`, []string{"ls;", "rm", "-rf", "/"}},
		{`echo $(whoami)`, []string{"echo", "$(whoami)"}},
	}
	for _, tt := range tests {
		got, err := Tokenize(tt.raw)
		if err != nil {
			t.Fatalf("Tokenize(%q) error: %v", tt.raw, err)
		}
		if len(got) != len(tt.want) {
			t.Fatalf("Tokenize(%q) = %v, want %v", tt.raw, got, tt.want)
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Fatalf("Tokenize(%q)[%d] = %q, want %q", tt.raw, i, got[i], tt.want[i])
			}
		}
	}
}

func TestEvaluate(t *testing.T) {
	alias := &config.Alias{
		Name: "test",
		Allow: []config.Allow{
			{Command: "borg serve"},
			{Command: "git-upload-pack", Args: []config.ArgConstraint{{Pos: 1, Glob: "/srv/git/*"}}, NumArgs: ptr(1)},
			{Command: "rsync", AnyArg: ptr("/srv/backup/*"), NoArg: ptr("/srv/backup/.secret/*")},
			{Command: "nc", Args: []config.ArgConstraint{{Pos: -1, Glob: "22"}}, NumArgs: ptr(2)},
		},
	}

	allowed := [][]string{
		{"borg", "serve"},
		{"borg", "serve", "--anything"}, // prefix match, any trailing args
		{"git-upload-pack", "/srv/git/repo.git"},
		{"rsync", "--server", ".", "/srv/backup/host"},
		{"nc", "host.internal", "22"},
	}
	for _, tokens := range allowed {
		if _, err := Evaluate(alias, tokens); err != nil {
			t.Errorf("Evaluate(%v) = denied (%v), want allowed", tokens, err)
		}
	}

	denied := [][]string{
		{"rm", "-rf", "/"},                              // unknown command
		{"git-upload-pack", "/srv/secret/repo.git"},     // glob mismatch
		{"git-upload-pack", "/srv/git/a", "/srv/git/b"}, // num_args
		{"rsync", "--server", ".", "/tmp/x"},            // any_arg unmet
		{"rsync", ".", "/srv/backup/.secret/keys"},      // no_arg narrowing
		{"nc", "host.internal", "2222"},                 // last arg != 22
		{"nc", "host.internal", "22", "extra"},          // num_args
	}
	for _, tokens := range denied {
		if _, err := Evaluate(alias, tokens); !errors.Is(err, ErrDenied) {
			t.Errorf("Evaluate(%v) = %v, want ErrDenied", tokens, err)
		}
	}
}

func TestResolveIndex(t *testing.T) {
	tests := []struct {
		pos, n  int
		wantIdx int
		wantOK  bool
	}{
		{1, 3, 0, true},
		{3, 3, 2, true},
		{4, 3, 3, false},
		{-1, 3, 2, true},
		{-3, 3, 0, true},
		{-4, 3, -1, false},
		{0, 3, 0, false},
	}
	for _, tt := range tests {
		idx, ok := resolveIndex(tt.pos, tt.n)
		if ok != tt.wantOK || (ok && idx != tt.wantIdx) {
			t.Errorf("resolveIndex(%d,%d) = (%d,%v), want (%d,%v)", tt.pos, tt.n, idx, ok, tt.wantIdx, tt.wantOK)
		}
	}
}
