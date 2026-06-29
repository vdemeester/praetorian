package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// resolveConfigPath should discover a JSON config (Nix-generated) in addition to
// the human-written HCL config, at both the XDG and /etc locations.
func TestResolveConfigPath_JSONFallback(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := filepath.Join(dir, ".config", "praetorian")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}

	jsonPath := filepath.Join(cfgDir, "config.json")
	if err := os.WriteFile(jsonPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := resolveConfigPath("")
	if err != nil {
		t.Fatalf("resolveConfigPath: %v", err)
	}
	if got != jsonPath {
		t.Errorf("expected JSON config %q, got %q", jsonPath, got)
	}
}

// When both config.hcl and config.json exist in the same directory, the
// human-written HCL takes precedence.
func TestResolveConfigPath_HCLBeatsJSON(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	cfgDir := filepath.Join(dir, ".config", "praetorian")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatal(err)
	}

	hclPath := filepath.Join(cfgDir, "config.hcl")
	jsonPath := filepath.Join(cfgDir, "config.json")
	for _, p := range []string{hclPath, jsonPath} {
		if err := os.WriteFile(p, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	got, err := resolveConfigPath("")
	if err != nil {
		t.Fatalf("resolveConfigPath: %v", err)
	}
	if got != hclPath {
		t.Errorf("expected HCL to win (%q), got %q", hclPath, got)
	}
}

// An explicit path always wins, even if missing.
func TestResolveConfigPath_ExplicitWins(t *testing.T) {
	got, err := resolveConfigPath("/nonexistent/path.hcl")
	if err != nil {
		t.Fatalf("resolveConfigPath: %v", err)
	}
	if got != "/nonexistent/path.hcl" {
		t.Errorf("expected explicit path, got %q", got)
	}
}
