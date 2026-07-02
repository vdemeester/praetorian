package engine

import (
	"path"
	"strings"
)

// globMatch reports whether name matches pattern. It extends path.Match with a
// recursive "**" segment that matches zero or more path segments. Every other
// segment is matched against exactly one path segment via path.Match, so "*",
// "?" and "[...]" keep their per-segment meaning and do not cross "/".
//
// Patterns without "**" behave exactly like path.Match.
func globMatch(pattern, name string) (bool, error) {
	if !strings.Contains(pattern, "**") {
		return path.Match(pattern, name)
	}
	return matchSegments(strings.Split(pattern, "/"), strings.Split(name, "/"))
}

// matchSegments matches pattern segments against name segments, treating a
// literal "**" segment as zero or more name segments.
func matchSegments(pat, name []string) (bool, error) {
	switch {
	case len(pat) == 0:
		return len(name) == 0, nil
	case pat[0] == "**":
		// Collapse consecutive "**" segments.
		if len(pat) > 1 && pat[1] == "**" {
			return matchSegments(pat[1:], name)
		}
		// "**" matches zero or more segments: try each possible consumption.
		for i := 0; i <= len(name); i++ {
			ok, err := matchSegments(pat[1:], name[i:])
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	case len(name) == 0:
		return false, nil
	default:
		ok, err := path.Match(pat[0], name[0])
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
		return matchSegments(pat[1:], name[1:])
	}
}

// matchArg matches a command argument against a glob pattern, with support for
// git's relative-path shorthand.
//
// Security model: a relative arg is resolved against the account home and
// path.Clean'd, then it MUST remain lexically inside the pattern's literal base
// (the leading path up to the first wildcard). This rejects "../" traversal
// that would escape the intended directory before any match is attempted.
// Absolute args are matched as-is; an escape simply fails to match the pattern.
func matchArg(pattern, arg, home string) (bool, error) {
	// Fast path: raw arg matches directly (covers absolute paths and
	// non-path args like "22").
	ok, err := globMatch(pattern, arg)
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}

	// Only attempt home resolution for a relative arg against an absolute,
	// path-shaped pattern.
	if strings.HasPrefix(arg, "/") || !strings.HasPrefix(pattern, "/") || home == "" {
		return false, nil
	}

	resolved := path.Clean(path.Join(home, arg))

	// Enforce base containment: the resolved path must stay within the
	// pattern's literal base directory.
	base := patternBase(pattern)
	if !withinBase(resolved, base) {
		return false, nil
	}

	return globMatch(pattern, resolved)
}

// patternBase returns the leading, wildcard-free directory prefix of a pattern.
// For "/home/vincent/git/**" it returns "/home/vincent/git".
func patternBase(pattern string) string {
	segs := strings.Split(pattern, "/")
	var base []string
	for _, s := range segs {
		if strings.ContainsAny(s, "*?[") {
			break
		}
		base = append(base, s)
	}
	return strings.Join(base, "/")
}

// withinBase reports whether p is lexically equal to or nested under base.
func withinBase(p, base string) bool {
	if base == "" {
		return true
	}
	if p == base {
		return true
	}
	return strings.HasPrefix(p, base+"/")
}
