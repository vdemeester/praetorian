package engine

import (
	"testing"
)

// TestGlobMatch covers the recursive matcher directly, without normalization.
func TestGlobMatch(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    bool
	}{
		// backward-compatible single-segment semantics: * does not cross /
		{"/srv/git/*", "/srv/git/repo.git", true},
		{"/srv/git/*", "/srv/git/public/home.git", false},
		{"/srv/git/*.git", "/srv/git/repo.git", true},
		{"/srv/git/*.git", "/srv/git/repo.txt", false},

		// ** matches zero or more segments
		{"/home/vincent/git/**", "/home/vincent/git/passage.git", true},
		{"/home/vincent/git/**", "/home/vincent/git/public/home.git", true},
		{"/home/vincent/git/**", "/home/vincent/git", true}, // zero segments
		{"/home/vincent/git/**", "/home/other/git/x.git", false},

		// ** in the middle
		{"/srv/**/repo.git", "/srv/a/b/repo.git", true},
		{"/srv/**/repo.git", "/srv/repo.git", true},
		{"/srv/**/repo.git", "/srv/a/b/other.git", false},

		// ? and [...] still work per segment
		{"/srv/git/repo?.git", "/srv/git/repo1.git", true},
		{"/srv/git/repo[0-9].git", "/srv/git/repo5.git", true},
		{"/srv/git/repo[0-9].git", "/srv/git/repoX.git", false},

		// consecutive ** collapse
		{"/srv/**/**/repo.git", "/srv/a/b/repo.git", true},
		{"/srv/**/**/repo.git", "/srv/repo.git", true},

		// non-path globs unaffected
		{"22", "22", true},
		{"22", "2222", false},
	}
	for _, tt := range tests {
		got, err := globMatch(tt.pattern, tt.name)
		if err != nil {
			t.Fatalf("globMatch(%q, %q) error: %v", tt.pattern, tt.name, err)
		}
		if got != tt.want {
			t.Errorf("globMatch(%q, %q) = %v, want %v", tt.pattern, tt.name, got, tt.want)
		}
	}
}

// TestMatchArgEdge covers empty-base patterns and error propagation.
func TestMatchArgEdge(t *testing.T) {
	// Pattern whose literal base is empty ("/**") accepts any relative arg
	// resolved under home.
	ok, err := matchArg("/**", "git/passage.git", "/home/vincent")
	if err != nil || !ok {
		t.Errorf(`matchArg("/**", ...) = (%v, %v), want (true, nil)`, ok, err)
	}
	// Bad pattern surfaces the error.
	if _, err := matchArg("/srv/[", "x", "/home/vincent"); err == nil {
		t.Error("matchArg with bad pattern: expected error, got nil")
	}
}

// TestMatchArg covers relative-path normalization against a home directory.
func TestMatchArg(t *testing.T) {
	const home = "/home/vincent"

	allowed := []struct{ pattern, arg string }{
		// absolute nested via **
		{"/home/vincent/git/**", "/home/vincent/git/public/home.git"},
		// relative shorthand resolved against home
		{"/home/vincent/git/**", "git/passage.git"},
		{"/home/vincent/git/**", "git/public/home.git"},
		// relative that cleans to a path still inside base
		{"/home/vincent/git/**", "git/./passage.git"},
		// non-path arg matched raw
		{"22", "22"},
	}
	for _, tt := range allowed {
		ok, err := matchArg(tt.pattern, tt.arg, home)
		if err != nil {
			t.Fatalf("matchArg(%q, %q) error: %v", tt.pattern, tt.arg, err)
		}
		if !ok {
			t.Errorf("matchArg(%q, %q) = false, want true", tt.pattern, tt.arg)
		}
	}

	denied := []struct{ pattern, arg string }{
		// traversal escaping the base must be denied
		{"/home/vincent/git/**", "git/../../etc/passwd"},
		{"/home/vincent/git/**", "../../etc/shadow"},
		{"/home/vincent/git/**", "git/../.ssh/authorized_keys"},
		// absolute escape simply does not match
		{"/home/vincent/git/**", "/etc/passwd"},
		// wrong base
		{"/home/vincent/git/**", "/home/other/git/x.git"},
	}
	for _, tt := range denied {
		ok, err := matchArg(tt.pattern, tt.arg, home)
		if err != nil {
			t.Fatalf("matchArg(%q, %q) error: %v", tt.pattern, tt.arg, err)
		}
		if ok {
			t.Errorf("matchArg(%q, %q) = true, want false (denied)", tt.pattern, tt.arg)
		}
	}
}
