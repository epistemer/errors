# Review 0001: the design, before the lock

| | |
|---|---|
| Date | 2026-11-14 |
| Under review | [interface.md](../interface.md) (draft), [ADR 0001](../adr/0001-two-messages.md), and a scratch spike of [brief.md](../brief.md) §8's sketches that passes §7 against i18n `v0.2.0` |
| Reviewer | Gemini 4 Hydrogen (High) through agy, about 70k tokens. Argon was cooling down for 15 hours |
| Triage | Claude |

Twelve findings: six accepted (one in part), two accepted as notes in the
docs, four rejected. Every accepted change to the interface is listed in its
§10 and was checked in the spike, where §7's tests still pass unchanged
along with a test for each change.

| # | Finding | Severity | Triage |
|---|---------|----------|--------|
| 1 | A nil pointer in an `error` or `i18n.Ter` (`var e *MyErr`) passed to `T`, `Public`, `KindOf`, `Wrap` or `With` gets its methods called and can panic | Critical | **Accepted.** A nil pointer counts as nil everywhere (§5.1, §5.3) |
| 2 | `Def1`–`Def3` implement `i18n.Ter`. One returned without `.With` is public, and people see its raw verbs. The spike confirmed it: a bare `ErrTooHigh` showed `يجب ألا يتجاوز %[1]s %[2]d`. Its `T(ctx, args ...any)` isn't type-checked either, which undercuts D5 | Critical | **Accepted; epistemer to confirm.** `Def1`–`Def3` get a typed `T(ctx, a A, …)` and aren't `Ter`s. A bare one isn't public: people see `Unknown`, logs show the verbs. §6.5's hand-written public error compiles unchanged. Lost: a `Def1`–`Def3` as an `i18n.Ter` label, which only ever rendered its verbs |
| 3 | `Join` of a private 5xx and a public 4xx hides the 5xx from §6.3's logging; `All` yields nothing for a tree with no public error | Major | **Accepted as notes** in §6.3. The status following what people see is D10, unchanged |
| 4 | A cycle in the tree recurses forever | Major | **Rejected.** `errors.Is` and `errors.As` assume no cycles too; a cycle is the wrapper's bug. Stated in §5.1 |
| 5 | `With(nil)` on a `Key1[error]` shows people `"payment declined: "`, not `Unknown`; a nil `any` argument becomes `""` and can print `%!d(string=)` | Major | **Accepted in part.** A nil in an `error` parameter fills its slot with `Unknown`, as an error with no public part does (§5.3). Other nils stay `""`: a nil `any` is the caller's choice |
| 6 | `Key(…, NotFound, Conflict)` silently keeps the last kind; `Kind(99)` is accepted | Major | **Accepted.** A repeated kind or one outside the eight panics (§5.8) |
| 7 | `errors.Unwrap(d.Wrap(cause))` is nil; `Public` ignores custom `As` methods | Major | **Accepted as a note** for `Unwrap` (§5.5), as for `errors.Join`. `As` was already §5.1 |
| 8 | `reflect.Type.String()` in meta can't tell two packages' `model.Item` apart; empty keys | Major | **Rejected.** Meta only lets `Load` refuse two meanings of one key; the collision makes it miss one conflict in an unlikely case and breaks nothing. i18n's `Load` already refuses an empty key |
| 9 | `Unknown` is a mutable variable; `KindOf(nil)` vs `Public(nil)` | Minor | **Accepted in part:** §4 now says `KindOf(nil)` is `Internal` (it was in §5.7). `Unknown` stays a variable, like `io.EOF` |
| 10 | An error passed to a plain i18n key's `T` prints `Error()` | Minor | **Accepted** as a rule in §5.3 |
| 11 | Allocations in argument preparation; `Public` and `KindOf` walk twice | Minor | **Rejected** for now: implementation, not contract. Revisit with a benchmark if a profile asks |
| 12 | The spike lacks `Key3`; `%T` on an error argument prints `string` | Minor | **Rejected.** The spike is scratch; `Key3` is M1.3. `%T` on a prepared argument is what it is |
