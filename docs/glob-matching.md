# Glob matching: recursive (`**`) and relative-path support

Tracks issue #29. Records the option assessment for how praetorian's engine
matches argument globs, and the security-critical normalization rules.

## Problem

The engine matched argument globs with the Go stdlib `path.Match`. Two gaps:

1. **`*` does not cross `/`.** `path.Match` treats `*` as "any run of
   non-separator characters", so `/home/vincent/git/*` does not match a nested
   repo like `/home/vincent/git/public/home.git`. There is no `**`.
2. **Relative vs absolute mismatch.** git's `host:path` shorthand sends a path
   relative to `$HOME` (`git-upload-pack git/passage.git`), but a configured
   glob is absolute (`/home/vincent/git/**`). Different strings → denied.

## Options considered

### A. New dependency: `github.com/bmatcuk/doublestar/v4`

- **Pros:** mature, well-tested, correct `**` semantics, widely used.
- **Cons:** adds a third-party dependency to a *security-sensitive core*. The
  whole point of praetorian is a small, auditable gate; every dependency is
  attack surface and supply-chain risk. `AGENTS.md` explicitly says: keep
  dependencies minimal, justify any new one.

### B. Go 1.26 stdlib

- Checked `path.Match` / `filepath.Match` on go1.26: **no `**` support**. `*`
  still stops at the separator by definition. There is no new stdlib glob API
  that covers recursive matching.
- **Verdict:** stdlib alone cannot satisfy the requirement.

### C. Hand-rolled matcher (chosen)

A small, auditable matcher tailored to our needs: segment-wise matching that
delegates each non-`**` segment to `path.Match` (preserving existing `*`, `?`,
`[...]` semantics) and treats `**` as "zero or more path segments".

- **Pros:** no new dependency; every line is auditable; behaviour is exactly
  what a security tool needs and nothing more; reuses battle-tested
  `path.Match` for the per-segment character classes.
- **Cons:** we own the correctness. Mitigated by strict TDD and adversarial
  tests (traversal, escapes).

**Decision: Option C.** For a default-deny security gate, auditability and a
minimal dependency surface outweigh the convenience of `doublestar`. The
recursive semantics we need (`base/**`) are narrow and testable.

## Semantics

### Recursive glob (`**`)

- The pattern and candidate are split on `/`.
- A literal `**` segment matches zero or more candidate segments.
- Every other segment is matched against one candidate segment with
  `path.Match` (so `*`, `?`, `[...]` keep their existing per-segment meaning
  and still do **not** cross `/`).
- This is backward compatible: patterns without `**` behave exactly as before.

### Relative-path normalization (security-critical)

Applied only when the candidate argument is **relative** (does not start with
`/`) and the pattern is an **absolute** path glob (starts with `/`):

1. Resolve the relative arg against the account home directory.
2. `path.Clean` the result.
3. Derive the pattern's **literal base** (the leading path up to the first
   wildcard segment). Reject — deny — any normalized path that is not
   lexically contained within that base. This blocks `..` traversal that
   escapes the intended directory (`git/../../etc/passwd`, etc.) *before* any
   glob match is attempted.
4. Match the normalized path against the pattern.

Absolute candidate args are matched as-is (no home resolution); traversal that
escapes the pattern base simply fails to match and is denied.

Adversarial cases covered by tests: `git/../../etc/passwd`, `../../etc/shadow`,
absolute escapes, and `..` that stays inside the base (allowed only if it still
matches).

## Postscript: differential comparison against `doublestar`

Out of curiosity, after implementing Option C the hand-rolled `globMatch` was
compared against `github.com/bmatcuk/doublestar/v4` (`doublestar.Match`) on 24
cases, including edge cases:

| Case                                      | Ours  | doublestar |
| ----------------------------------------- | ----- | ---------- |
| `/srv/git/*` vs nested `.../public/x.git` | false | false      |
| `/home/vincent/git/**` vs nested repo     | true  | true       |
| `/home/vincent/git/**` vs base (0 segs)   | true  | true       |
| `/srv/**/repo.git` (mid `**`)             | true  | true       |
| `/srv/**/**/repo.git` (consecutive `**`)  | true  | true       |
| `repo[0-9].git` char classes              | =     | =          |
| `/srv/x**y/repo.git` (intra-segment `**`) | true  | true       |
| `/srv/**.git` vs `/srv/a/b.git`           | false | false      |
| `**`, `**/repo.git`, `/a/**/b/**/c`       | =     | =          |

**Result: 24/24 identical**, including the subtle rule that `**` is only
recursive when it is a *whole path segment* — intra-segment `**` degrades to
ordinary per-segment matching in both implementations.

Conclusion: the hand-rolled matcher is behaviourally equivalent to `doublestar`
for praetorian's patterns, while adding zero dependencies to the security core.
The differential test was a throwaway module outside the repo; it confirmed the
choice rather than becoming a permanent dependency.
