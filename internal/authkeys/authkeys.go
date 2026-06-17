// Package authkeys analyzes OpenSSH authorized_keys files and cross-checks
// praetorian command= directives against a loaded config.
package authkeys

import (
	"bufio"
	"io"
	"strings"

	"github.com/google/shlex"
	"github.com/vdemeester/praetorian/internal/config"
)

// Kind classifies an authorized_keys entry relative to praetorian config.
type Kind int

const (
	// PraetorianFound means command="praetorian run <alias>" and the alias is in config.
	PraetorianFound Kind = iota
	// PraetorianMissing means the entry references a praetorian alias absent from config.
	PraetorianMissing
	// Unrestricted means there is no command= directive (e.g. FIDO2 keys).
	Unrestricted
	// OtherCommand means a command= directive that is not praetorian.
	OtherCommand
)

// Entry is the analysis result for a single authorized_keys line.
type Entry struct {
	Kind       Kind
	Alias      string // praetorian alias (PraetorianFound/Missing)
	AllowCount int    // number of allow rules (PraetorianFound)
	Command    string // raw command= value (OtherCommand)
}

// Analyze parses authorized_keys lines and classifies each against cfg.
func Analyze(r io.Reader, cfg *config.Config) ([]Entry, error) {
	var entries []Entry
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		entries = append(entries, classify(line, cfg))
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func classify(line string, cfg *config.Config) Entry {
	command, ok := extractCommand(line)
	if !ok {
		return Entry{Kind: Unrestricted}
	}
	alias, ok := praetorianAlias(command)
	if !ok {
		return Entry{Kind: OtherCommand, Command: command}
	}
	if a := cfg.Lookup(alias); a != nil {
		return Entry{Kind: PraetorianFound, Alias: alias, AllowCount: len(a.Allow)}
	}
	return Entry{Kind: PraetorianMissing, Alias: alias}
}

// extractCommand returns the command="..." option value, if present.
func extractCommand(line string) (string, bool) {
	const key = `command="`
	i := strings.Index(line, key)
	if i < 0 {
		return "", false
	}
	rest := line[i+len(key):]
	// Find the closing quote, honoring backslash escapes.
	var sb strings.Builder
	for j := 0; j < len(rest); j++ {
		c := rest[j]
		if c == '\\' && j+1 < len(rest) {
			j++
			sb.WriteByte(rest[j])
			continue
		}
		if c == '"' {
			return sb.String(), true
		}
		sb.WriteByte(c)
	}
	return "", false
}

// praetorianAlias returns the alias from a "praetorian run <alias>" command.
func praetorianAlias(command string) (string, bool) {
	tokens, err := shlex.Split(command)
	if err != nil || len(tokens) < 3 {
		return "", false
	}
	if base(tokens[0]) != "praetorian" || tokens[1] != "run" {
		return "", false
	}
	// Skip flags (e.g. --config PATH) to find the positional alias.
	for i := 2; i < len(tokens); i++ {
		if strings.HasPrefix(tokens[i], "-") {
			if tokens[i] == "--config" {
				i++ // skip its value
			}
			continue
		}
		return tokens[i], true
	}
	return "", false
}

func base(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
