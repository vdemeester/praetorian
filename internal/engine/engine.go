// Package engine tokenizes and evaluates SSH commands against allow rules.
//
// Security model: structural, not validation-based. The command string is
// tokenized with a shell-like lexer (google/shlex) and the matched executable
// is run via syscall.Exec — no shell ever interprets it. $(), backticks, ;, |
// are inert bytes. There is no TOCTOU gap: what is validated is exactly what is
// executed.
package engine

import (
	"errors"
	"fmt"
	"path"

	"github.com/google/shlex"
	"github.com/vdemeester/praetorian/internal/config"
)

// ErrDenied is returned when no allow rule matches the command.
var ErrDenied = errors.New("denied")

// Tokenize splits a raw SSH_ORIGINAL_COMMAND into tokens using shell-like
// quoting rules, without invoking a shell.
func Tokenize(raw string) ([]string, error) {
	tokens, err := shlex.Split(raw)
	if err != nil {
		return nil, fmt.Errorf("tokenizing command: %w", err)
	}
	return tokens, nil
}

// Evaluate checks tokens against an alias's allow rules. On success it returns
// the matched rule. On failure it returns ErrDenied (wrapped with a reason).
func Evaluate(alias *config.Alias, tokens []string) (*config.Allow, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("%w: empty command", ErrDenied)
	}
	var lastErr error
	for i := range alias.Allow {
		rule := &alias.Allow[i]
		if err := matchRule(rule, tokens); err != nil {
			lastErr = err
			continue
		}
		return rule, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("%w: no matching allow rule", ErrDenied)
	}
	return nil, lastErr
}

// matchRule reports whether tokens satisfy a single allow rule.
func matchRule(rule *config.Allow, tokens []string) error {
	prefix, err := shlex.Split(rule.Command)
	if err != nil {
		return fmt.Errorf("invalid allow command %q: %w", rule.Command, err)
	}
	if len(prefix) == 0 {
		return fmt.Errorf("invalid allow command %q: empty", rule.Command)
	}
	if len(tokens) < len(prefix) {
		return fmt.Errorf("%w: prefix mismatch", ErrDenied)
	}
	for i, p := range prefix {
		if tokens[i] != p {
			return fmt.Errorf("%w: prefix mismatch", ErrDenied)
		}
	}
	args := tokens[len(prefix):]

	if rule.NumArgs != nil && len(args) != *rule.NumArgs {
		return fmt.Errorf("%w: expected %d args, got %d", ErrDenied, *rule.NumArgs, len(args))
	}
	for _, ac := range rule.Args {
		idx, ok := resolveIndex(ac.Pos, len(args))
		if !ok {
			return fmt.Errorf("%w: arg position %d out of range", ErrDenied, ac.Pos)
		}
		match, err := path.Match(ac.Glob, args[idx])
		if err != nil {
			return fmt.Errorf("bad glob %q: %w", ac.Glob, err)
		}
		if !match {
			return fmt.Errorf("%w: arg %d (%q) does not match %q", ErrDenied, ac.Pos, args[idx], ac.Glob)
		}
	}
	if rule.AnyArg != nil {
		found := false
		for _, a := range args {
			if m, err := path.Match(*rule.AnyArg, a); err != nil {
				return fmt.Errorf("bad glob %q: %w", *rule.AnyArg, err)
			} else if m {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("%w: no arg matches any_arg %q", ErrDenied, *rule.AnyArg)
		}
	}
	if rule.NoArg != nil {
		for _, a := range args {
			if m, err := path.Match(*rule.NoArg, a); err != nil {
				return fmt.Errorf("bad glob %q: %w", *rule.NoArg, err)
			} else if m {
				return fmt.Errorf("%w: arg %q matches forbidden no_arg %q", ErrDenied, a, *rule.NoArg)
			}
		}
	}
	return nil
}

// resolveIndex converts a 1-based position (negative = from end) to a 0-based
// slice index, reporting whether it is in range.
func resolveIndex(pos, n int) (int, bool) {
	switch {
	case pos > 0:
		idx := pos - 1
		return idx, idx < n
	case pos < 0:
		idx := n + pos
		return idx, idx >= 0
	default:
		return 0, false // pos == 0 is invalid (1-based)
	}
}
