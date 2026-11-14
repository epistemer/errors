# Status log

Newest first. Every working session adds an entry: what landed, the decisions
taken for you, what's next, and what's waiting on you.

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
- **Next:** M1, once the interface is locked.
- **Waiting on you:**
  - Locking [interface.md](interface.md) (M0.2's done-when), with or
    without review finding 2.
  - Delegating ownership on the terms of [ownership.md](ownership.md).
