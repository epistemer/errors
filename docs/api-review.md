# API review for v0.1.0 (M3.2)

Every exported identifier of package `errors`, as `go doc -all` lists it on
2026-11-14, against [interface.md](interface.md) §4 and the rule it serves.
"Checked" names the test that pins it down. Reviewed by Claude.

## The standard library's API (§4)

| Identifier | Interface | Checked |
|---|---|---|
| `ErrUnsupported` | the standard library's | `TestStandardAPI`: the same value |
| `New`, `Is`, `As`, `AsType`, `Unwrap`, `Join` | the standard library's | `TestStandardAPI`: each against the standard library's, `AsType` found and not found, `Join` of nothing |

## Declaring (§4, §5.8, §5.9)

| Identifier | Interface | Checked |
|---|---|---|
| `Key` | signature as §4; text never formatted | `TestRules/Key's_text_is_never_formatted`; §7 |
| `Key1`, `Key2`, `Key3` | signatures as §4; the checks of §5.8 | `TestCheckText` (§7's 14 cases); `TestReviewChanges/Key3` |
| kinds as options | §5.8: a repeated or unknown kind panics | `TestDeclareRefusesBadKinds` |
| meta | §5.9: stable bytes, none for a plain `Key`, `Load` refuses two meanings | `TestMeta`; `internal/conflicttest` |
| `Option` | an interface with an unexported method; `Kind` is one | `TestKinds` (compile-time assignment) |
| `Def`: `Error`, `K`, `T`, `Kind`, `Wrap` | as §4; `Wrap(nil)` is `d`, also for a nil pointer as the cause | §7 `TestShow`; `TestRules/accessors`; `TestReviewChanges/nil_and_nil_pointers` |
| `Def1`–`Def3`: `Error`, `K`, `T`, `Kind`, `With`, `Wrap` | as §4 with review 0001's typed `T`; not an `i18n.Ter` | `TestReviewChanges` (bare declarations, typed `T`, `Key3`); `TestRules/accessors`, `TestRules/Wrap_without_a_cause` |

## Showing (§4, §5.1–§5.7)

| Identifier | Interface | Checked |
|---|---|---|
| `Unknown` | `Key("errors.unknown", "something went wrong", Internal)` | the declaration itself; §7 `TestShow` (code `errors.unknown`) |
| `Public` | §5.1: first public error in pre-order, `Unknown`, nil for nil; a keyless one has key `""` | §7 `TestShow`, `TestNil`; `TestRules/Public` |
| `T` | `Public(err).T(ctx)`, `""` for nil | §7 `TestShow`, `TestAll`, `TestNil`, `TestWhiteLabel` |
| `All` | §5.1: outermost public error per branch; first is `Public` | §7 `TestAll`; `TestRules/All_stops_when_asked` |
| `KindOf` | §5.7: own kind, else first error argument's, else `Internal`; `Internal` for nil; `Wrap` lends nothing | §7 `TestShow`; `TestRules/KindOf` |
| occurrences (unexported) | §4: `error`, `i18n.TKer`, `Kind()`, `Is`, `Unwrap() []error`; §5.2, §5.3, §5.6 | §7 `TestIdentity`, `TestShow`, `TestNoPrivateTextReachesPeople`; `TestErrorIgnoresEdits`; `TestRules/identity`, `TestRules/Unwrap` |

## Kinds (§4, §5.7)

| Identifier | Interface | Checked |
|---|---|---|
| `Kind` and its eight constants | values, names and statuses of §5.7's table | `TestKinds`, every kind and the zero `Kind` |
| `Kind.Status`, `Kind.String` | 500 and `"unspecified"` for the zero `Kind`; `Kind(n)` and 500 outside the eight | `TestKinds` |

## Findings

1. **No identifier outside §4.** `go doc` lists exactly §4's API; nothing
   else is exported.
2. **`Kind.String` of a value outside the eight** returns `Kind(n)`. §4
   doesn't say; such a value can't come from a declaration (§5.8) and only
   from a conversion. Left as is.
3. **`Def1`–`Def3`'s `Wrap(nil, …)`** returns an occurrence, equal in every
   behaviour to `With(…)`. §4 only defines `Def.Wrap(nil)`. Left as is.
4. **`Unknown` is a variable**, as §4 declares it. Review 0001 finding 9
   kept it so; reassigning it is the app's bug, like reassigning `io.EOF`.

None needs a change or an amendment. A cross-model review of this
document, the README and the examples (Gemini 3.8 Flash High) found claims
the tables didn't support; the tests named in them were added.
