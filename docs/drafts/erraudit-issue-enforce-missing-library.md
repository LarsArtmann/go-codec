# Draft — erraudit issue (own repo, solicited; NOT filed)

Target repo: `github.com/larsartmann/erraudit`
Proposed title: Warn when `--enforce-go-error-family`/`--enforce-samber-oops` names a library absent from go.mod

---

## Symptom

Both flags are pure boolean pass-throughs
(`cmd/erraudit/cmd/appconfig.go:35-36` → `internal/portfolio/audit_args.go:59-63`):
nothing checks that the audited module actually depends on the named library.
Running `--enforce-samber-oops` on a repo that neither imports nor documents
samber/oops produces a confident report flagging every `errors.New`/`fmt.Errorf`
as a violation — a misleading audit, not a wrong-but-obvious one. That exact
trap produced the misleading input report in the 2026-09-11 go-codec session
(archived status report §e-6).

Verified against erraudit `v0.6.1-6-g7773253`.

## Root cause (source)

```go
EnforceErrorFamily bool `default:"false" flag:"enforce-go-error-family" ...`
EnforceSamberOops  bool `default:"false" flag:"enforce-samber-oops"     ...`
// cmd/erraudit/cmd/appconfig.go:35-36

appendBool("--enforce-go-error-family", cfg.EnforceErrorFamily)
appendBool("--enforce-samber-oops", cfg.EnforceSamberOops)
// internal/portfolio/audit_args.go:60-61
```

The analyzer trusts the flag unconditionally
(`internal/ast/analyzer.go:195-206`); go.mod is already parsed for the module
graph, so the dependency facts are available at decision time.

## Design

- [ ] Warn (stderr, non-fatal) when `--enforce-go-error-family` is set but
      `github.com/larsartmann/go-error-family` is absent from the module's
      requirements — same for `--enforce-samber-oops` and `github.com/samber/oops`
- [ ] Alternative/complement: auto-detect go-error-family from go.mod and
      enable enforcement implicitly (was 2026-09-11 follow-up #30)
- [ ] Warning text states what was checked, so a typo'd module path in go.mod
      does not silently downgrade to warn-only

## To verify

- Repro: `erraudit ./... --enforce-samber-oops` in a module without the
  dependency → warning emitted, report otherwise unchanged
- Same command in a module WITH the dependency → no warning

💘 Generated with Crush
