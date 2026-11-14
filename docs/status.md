# Status log

Newest first. Every working session adds an entry: what landed, the decisions
taken for you, what's next, and what's waiting on you.

## 2026-11-14: the package (M1, M2)

- **Landed** (PR #2): the whole of interface §4 and §5. `Key`–`Key3`, the
  declaration checks and meta; occurrences; the walk (`Public`, `T`,
  `All`); arguments and composition; `KindOf`; the standard library's API;
  kinds. Built from brief §8's sketches with review 0001's changes, each
  type's methods written out.
  - Tests: §7's acceptance tests verbatim (`acceptance_test.go`), the
    checker's 14 cases and meta (internal), the standard API and every kind,
    `Error()` unchanged by Arabic and English edits, review 0001's changes,
    and `TestRules`. `internal/conflicttest`, a test binary of its own,
    shows `Load` refusing one key with two kinds and one with two argument
    lists. Coverage about 91%.
  - [Review 0002](reviews/0002-m1-m2.md) by Gemini 4 Hydrogen (Medium),
    about 47k tokens (Argon cooling down): five findings, three accepted.
    The bug it found: a nil pointer of a type that is neither an error nor
    a label reached people as `%!s(*User=<nil>)`.
- **Decided for you:**
  - The checker reports an unused argument as unused (fmt's `%!(EXTRA …)`
    used to make it "can't format").
  - M1.1 and M1.2 done here, not by Gemini: the spike had written them.
    Delegating would have cost a worktree and a review for a dozen lines.
- **Next:** M3: examples, the README quickstart, the API review, `v0.1.0`,
  and the skill.
- **Waiting on you:** nothing.

## 2026-11-14: the repository starts (M0.1, M0.2 drafted)

- **Landed** (PR for M0):
  - The module (Go 1.26, MIT), `doc.go`, `CLAUDE.md`, README, and CI with
    the same 12 jobs as models.
  - [brief.md](brief.md), exported from the design brief in epistemer's
    Drive and cleaned of export escapes; content unchanged.
  - [ADR 0001](adr/0001-two-messages.md) from brief §§1–5,
    [interface.md](interface.md) from §§3–7 (section numbers kept, so the
    brief's cross-references hold), [roadmap.md](roadmap.md) from §12, and
    [ownership.md](ownership.md) on models' terms, marked proposed.
  - **A spike** in a scratch folder, not committed: brief §8's sketches
    filled out into a package (`Key3` left out), run against the real i18n
    `v0.2.0`. All of §7's acceptance tests (21 cases in 6 tests) and the 14
    checker cases pass with `-race`. The brief had only checked them against
    a stand-in.
  - [Review 0001](reviews/0001-design.md) of the interface, the ADR and the
    spike, by Gemini 4 Hydrogen (High), about 70k tokens (Argon cooling
    down for 15 hours): twelve findings, six accepted (one in part), two
    accepted as notes, four rejected. The accepted ones were made in
    interface.md before the lock (its §10) and checked in the spike.
- **Decided for you** (overturnable before the lock):
  - Review finding 2, the biggest: `Def1`–`Def3` take a typed
    `T(ctx, a A, …)` and are not `i18n.Ter`s. A declaration returned without
    `.With` then shows people `Unknown` instead of its raw verbs (the spike
    showed `يجب ألا يتجاوز %[1]s %[2]d` before the change).
  - A nil or nil pointer counts as nil; a nil `error` argument fills its
    slot with `Unknown`; a repeated or unknown `Kind` panics.
  - M0.1 done here rather than by Gemini: it was a copy of models' files.
- **Decided by you (2026-11-14):** the interface is locked, with the typed
  `T` of review finding 2 and the brief's (Claude) decisions as written;
  ownership delegated on models' terms.
- **Next:** M1.
- **Waiting on you:** nothing.
