# Review 0003: the release docs (M3.1–M3.3)

| | |
|---|---|
| Date | 2026-11-14 |
| Under review | The README quickstart, the godoc examples, [api-review.md](../api-review.md) and the CHANGELOG, against `go doc -all` |
| Reviewer | Gemini 3.8 Flash (High) through agy, about 223k tokens |
| Triage | Claude |

Eleven findings: ten accepted, one in part.

| # | Finding | Triage |
|---|---------|--------|
| 1 | The README names `ErrOrderNotSaved`, which it never declares | **Accepted:** "a declaration's `Wrap`" |
| 2 | `Public(nil).K()` panics; the README pairs it with nil-safe calls | **Accepted:** the README says "for a non-nil `err`", and that a keyless public error's code is `""` |
| 3 | The package doc said `All` takes the first public error | **Accepted:** reworded |
| 4 | `Unwrap`'s doc implied an occurrence with one cause unwraps | **Accepted:** it is nil for every occurrence; reworded |
| 5 | The API review claimed "a nil pointer is nil" for `Def`'s methods, which meant causes and arguments, not receivers | **Accepted:** reworded |
| 6–8 | The API review cited tests that didn't check `Def1.Wrap(nil, …)`, `Kind(99)`, or an occurrence's `Unwrap() []error` | **Accepted:** the tests were added (`TestRules`, `TestKinds`) |
| 9 | The CHANGELOG said `Key` has argument types and checks | **Accepted:** reworded |
| 10 | `ExampleKey1` showed `Def1.With`, not `Key1` | **Accepted:** renamed `ExampleDef1_With` |
| 11 | The quickstart's store sits at a fixed temp path; seeding only when `CreateLocale` succeeds runs against stale contents | **In part:** a fresh temp dir per run. Seeding only on creation is i18n's documented pattern and stays |
