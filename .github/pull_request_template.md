## Summary

<!-- What does this PR change and why? -->

## Changes

<!-- Bullet the notable changes. -->

-

## Testing

<!-- How was this verified? TDD is expected: tests added/updated for new behavior. -->

- [ ] `make check` passes (fmt + vet + lint + test)
- [ ] New/changed behavior is covered by tests (RED → GREEN)

## Checklist

- [ ] No shell is introduced into the validate/exec path (commands stay
      tokenized + `syscall.Exec`)
- [ ] Default-deny preserved; no new top-level deny rules
- [ ] User-facing denials remain terse; details only in logs
- [ ] No new dependencies (or justified above)
- [ ] Docs/README updated if behavior or config changed
