# errors — Roadmap to the first release

> Status: **accepted plan**, written 2026-11-14 from [brief.md](brief.md)
> §12. Module: `github.com/epistemer/errors`.
>
> | Document | Answers |
> |----------|---------|
> | [ADR 0001](adr/0001-two-messages.md) | *What* the package does, and why that way |
> | [interface.md](interface.md) | *What* an app writes and gets back (locked 2026-11-14) |
> | this file | *In what order*, and *when each piece counts as done* |

## How to use this document

- **Ownership.** Claude owns errors until `v1.0.0`, on models' terms
  ([ownership.md](ownership.md), delegated 2026-11-14).
- **IDs.** Every item has an ID (`M1.3`) and fits one PR or one focused
  session. Branch names, PR titles and commit subjects start with the ID.
- **Interface first.** M0 ends with the lock of `docs/interface.md`. After
  that a change to what an app writes or gets back is an amendment.
- **Versions.** The first usable release is `v0.1.0`. `v1.0.0` waits until
  zp's business code runs on it.
- **Agents.** Claude owns the cross-cutting core (M1.3, M1.4, M2.1–M2.3)
  and reviews Gemini's work; Gemini, through agy, takes leaf items with a
  crisp done-when (M1.1, M1.2, M3.1). Each delegated item runs in its own
  git worktree outside the repository.

## Overview

| Milestone | Goal | Needs |
|-----------|------|-------|
| [M0](#m0--foundations) | Module, CI, design record, interface locked | – |
| [M1](#m1--declaring) | The standard API, kinds, declarations, occurrences | M0 |
| [M2](#m2--showing) | The walk, arguments and composition, `KindOf` | M1 |
| [M3](#m3--release) | Examples, README, API review, `v0.1.0`, the skill | M2 |

## Definition of done (every work item)

1. `go test -race ./...` passes on the whole CI matrix (Go 1.26 and 1.27 on
   Linux, macOS and Windows).
2. It lands as a PR with green CI and a review by the other model family.
3. If the behaviour differs from interface.md, that document is amended in
   the same PR.
4. A new dependency needs an ADR. The module depends on i18n only.
5. Nothing of zp's names enters the repository: tests and examples use a toy
   shop.

---

## M0 — Foundations

| ✓ | ID | Item | Done when |
|---|----|------|-----------|
| ☑ | M0.1 | Module `github.com/epistemer/errors` (Go 1.26, MIT), `doc.go`, `CLAUDE.md`, and CI like i18n's | CI runs green on a PR |
| ☑ | M0.2 | ADR 0001 (brief §§1–5), `docs/interface.md` (brief §§3–7), this roadmap, `docs/status.md`, and the brief as `docs/brief.md` | epistemer locks the interface |

## M1 — Declaring

| ✓ | ID | Item | Done when |
|---|----|------|-----------|
| ☑ | M1.1 | The standard library's API (interface §4) | A test shows each one behaves as the standard library's, `AsType` included |
| ☑ | M1.2 | `Kind` and `Option`: names, statuses, `String` | A table test covers every kind and the zero `Kind` |
| ☑ | M1.3 | `Key`–`Key3`, the checks (§5.8) and meta (§5.9) | The checker's internal test covers §7's fourteen cases; meta is byte-stable and absent for a plain `Key`; a test package of its own shows `Load` refusing one key declared with two kinds |
| ☑ | M1.4 | Occurrences: `With`, `Wrap`, `Is`, `Unwrap`, `Error()` | `TestIdentity` passes, and `Error()` stays the code's text with an Arabic edit and an English edit in the store |

## M2 — Showing

| ✓ | ID | Item | Done when |
|---|----|------|-----------|
| ☑ | M2.1 | The walk: `Public`, `T`, `All` (§5.1) | `TestAll` and `TestNil` pass |
| ☑ | M2.2 | Arguments and composition (§5.3, §5.4) | `TestShow`, `TestWhiteLabel` and `TestNoPrivateTextReachesPeople` pass |
| ☑ | M2.3 | `KindOf` (§5.7) | `TestShow`'s kinds pass, including inheritance through two frames and a `Wrap` cause that lends nothing |

## M3 — Release

| ✓ | ID | Item | Done when |
|---|----|------|-----------|
| ☑ | M3.1 | Godoc examples, and a README with a quickstart | The examples run as tests, and the quickstart's output is pasted from a real run |
| ☑ | M3.2 | API review of every exported identifier against the interface | A written review lists each one and how it was checked, and its findings are fixed or filed |
| ☑ | M3.3 | CHANGELOG, the `v0.1.0` tag, and a global Claude Code skill (`~/.claude/skills/errors/`) like cfgx's, i18n's and genx's | `go get github.com/epistemer/errors@v0.1.0` works through the module proxy |

## Alongside

- **i18n amendment** (interface §9), in the i18n repository under its own
  terms, released as i18n `v0.3.0`. This package doesn't wait for it.

## After v0.1.0 (backlog)

Interface §8 lists what v1 leaves out: field paths, stack traces, an HTTP
writer, inner codes in API bodies, plural forms and locale-aware numbers in
arguments, and a vet check for a `Def1`–`Def3` returned without its
arguments.

## Risks

| Risk | Likelihood / impact | Mitigation |
|------|---------------------|------------|
| The acceptance tests were written against a stand-in for i18n | Low / medium | Retired on 2026-11-14: a scratch spike of brief §8's sketches passed all of §7 with `-race` against i18n `v0.2.0` |
| Shadowing the standard library's `errors` confuses tools or readers | Medium / low | The package re-exports the whole standard API (M1.1), so one import serves |
| A `Def1`–`Def3` returned without its arguments shows its verbs | Medium / low | Tests that check messages catch it; a vet check is in the backlog |
| The i18n amendment changes how labels render inside i18n's own `T` | Low / low | This package prepares its own arguments and doesn't depend on it |
