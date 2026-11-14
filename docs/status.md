# Status log

Newest first. Every working session adds an entry: what landed, the decisions
taken for you, what's next, and what's waiting on you.

## 2026-11-14: the i18n amendment (alongside)

- **Landed, in i18n:** the amendment of interface §9, approved by epistemer,
  released as i18n `v0.3.0` (epistemer/i18n PR #11): a `Ter` argument
  renders in the request's locale. errors' tests pass unchanged against
  `v0.3.0` with `-race`; go.mod stays on `v0.2.0`, the lowest it needs.
- **Skills:** the `errors` skill names both i18n versions and says a
  wrapped error passed to a plain i18n key still prints `Error()`; the
  `i18n` skill's note on errors was corrected to say the same.
- **Next:** nothing on the roadmap. `v1.0.0` waits for zp's business code
  to run on errors (brief §11), in zp's own session.
- **Waiting on you:** nothing.

## 2026-11-14: v0.1.0 (M3.3)

- **Landed:** PR #3 merged after CI passed on all 12 jobs; `v0.1.0` tagged
  on `d3760da` and pushed. `go get github.com/epistemer/errors@v0.1.0`
  resolves through proxy.golang.org, and a clean module builds and runs
  against it.
- **Decided for you:** nothing.
- **Next:** the i18n amendment (interface §9), in i18n's repository under
  its terms; its text comes back to you there. Then `v1.0.0` waits for zp's
  business code to run on it (brief §11).
- **Waiting on you:** nothing. GitHub auto-merge is off for this repository;
  merges wait for CI and are done by hand, which works.

## 2026-11-14: the release docs and the skill (M3)

- **Landed** (PR #3, merged):
  - Godoc examples that run as tests, and a README quickstart whose output
    is pasted from a real run in Arabic against i18n `v0.2.0`.
  - [api-review.md](api-review.md): every exported identifier against
    interface §4, with the test that checks it. No changes or amendments.
  - CHANGELOG for `v0.1.0`.
  - [Review 0003](reviews/0003-m3-release.md) by Gemini 3.8 Flash (High),
    about 223k tokens: eleven findings, ten accepted and one in part. They
    were doc wording and three missing tests, now added.
  - Outside the repository: the global skill `~/.claude/skills/errors/`,
    and errors added to epistemer's global CLAUDE.md beside the other four
    packages.
- **Next:** merge PR #3 once CI is green, tag `v0.1.0`, and check that
  `go get github.com/epistemer/errors@v0.1.0` resolves through the module
  proxy (M3.3). After that: the i18n amendment (interface §9), in i18n's
  repository.
- **Waiting on you:** nothing.

## 2026-11-14: the package (M1, M2)

- **Landed** (PR #2, merged): the whole of interface §4 and §5. `Key`–`Key3`, the
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
