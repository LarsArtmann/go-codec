# Draft — erraudit issue (own repo, solicited; NOT filed)

Target repo: `github.com/larsartmann/erraudit`
Proposed title: `--enforce-generic-return` misses go-error-family creators (`WrapOncef`, `WrapRejectionf`, …)

---

## Symptom

With `--enforce-generic-return`, a function that creates errors only via
go-error-family constructors is silently not flagged, while the identical
function using `errors.New` is:

```go
func controlCreator(err error) error {
	return errors.New("control: creates internally") // flagged
}

func wrapOncefCreator(err error) error {
	return errorfamily.WrapOncef(err, errorfamily.Rejection, "repro.boom", "wrapped") // NOT flagged
}
```

`erraudit ./... --enforce-generic-return` on that file: exactly 1 violation
(`controlCreator`, WARNING `generic_return`, exit 2). Same result with
`--type-aware`. Verified against erraudit `v0.6.1-6-g7773253`.

## Root cause (source)

`createsErrorInternally` (`internal/ast/predicates.go:95-127`) recognizes two
shapes only:

```go
return a.isPackageMethodCall(call, "errors", "New") ||
	a.isPackageMethodCall(call, "errors", "Join") ||
	a.isErrorfCall(call) // internal/ast/predicates.go:244-248

if method == methodWrap || method == methodWrapf { // :116, constants :22-23
```

The exact-name match hits `Wrap`/`Wrapf` but none of the other go-error-family
constructors: `WrapOnce(f)`, `WrapRejection(f)`, `WrapCorruption(f)`,
`WrapInfrastructure(f)`, `WrapOrchestration(f)`, `WrapTransient(f)`,
`WrapConflict(f)`. Those are the idiomatic creators in a go-error-family
codebase, so the flag under-reports exactly its target audience — the same
projects `--enforce-go-error-family` exists for.

## Design

- [ ] Recognize the go-error-family constructor set in `createsErrorInternally`
      (explicit name list, or prefix match on `Wrap` when the callee resolves
      to the go-error-family import)
- [ ] Cover non-`f` variants too (`WrapOnce`, `WrapRejection`, …)
- [ ] Keep exact-name `Wrap`/`Wrapf` matching for other packages (samber/oops)
      unchanged

## To verify

- Repro above: both functions flagged, exit 2
- `TestEnforceGenericReturn*` suite still green

💘 Generated with Crush
