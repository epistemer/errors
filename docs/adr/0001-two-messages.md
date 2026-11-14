# ADR 0001: Errors with two messages

| | |
|---|---|
| Status | Proposed (accepted in design on 2026-11-14; final when the interface is locked) |
| Date | 2026-11-14 |
| Roadmap item | M0.2 |
| Deciders | epistemer; proposed in a design session with Claude ([brief.md](../brief.md)) |

## Context

neozp needs errors that people can read in their own language, and that
developers can still wrap, test and log. zp, the system neozp replaces,
shows why the obvious design fails. Its numbers (counts only; zp is private):

- 454 `errors.New` sentinels, about half declared in templ files next to the
  forms that use them. A generator finds them in the AST and writes a
  registry that maps each message to an i18n key, deduplicated by message
  (about 400 remain). `errors.T(ctx, err)` looks `err.Error()` up in that
  registry.
- `Error()` is the developers' message and wrapping changes it, so a wrapped
  error rendered as "unknown error occurred". The code adapted by giving up
  wrapping:
  - the service layer has no `fmt.Errorf` at all, and about 280 sites log
    the cause and return a generic sentinel in its place;
  - the database layer logs the driver's error and returns a sentinel, so
    nothing above it can test for `context.Canceled`;
  - 80 API handlers return `err.Error()`, so clients get English, internals
    included, whatever language they ask for;
  - every handler picks its HTTP status with its own `errors.Is` checks (37
    in all).
- A stub that implemented both `error` and `i18n.Ter` failed from the other
  side: `fmt.Errorf("…: %w", e)` returns a `*fmt.wrapError`, which isn't a
  `Ter`.
- One message could never contain another, so "checkout failed: insufficient
  balance" couldn't be translated as a whole.

Constraints: the standard library's wrapping (`%w`, two `%w`s,
`errors.Join`) and third-party wrappers must keep working unchanged; i18n
(`v0.2.0`) is the only translation mechanism, its keys are permanent, and
deployments edit text in the running app (white-labeling).

## Options

1. **Translate `Error()`** (zp's way): look the developer's text up in a
   registry. Breaks as soon as anything wraps, needs a generator, and makes
   log text and user text the same string. Rejected: it is the problem.
2. **Translatable wrappers**: every wrapper implements `Ter` and builds the
   user's message up the chain. Every wrapper in every library would have to
   know about i18n, and `fmt.Errorf` never will. Rejected.
3. **Two messages, the user's picked from the tree** (chosen): `Error()`
   stays the code's text for logs; `T(ctx)` comes from the first error in
   the tree, in `errors.As` order, that implements `i18n.Ter`. Wrappers stay
   ignorant of i18n. Composition is explicit: a message with a slot takes
   the inner error as a typed argument.

For argument counts, three fixes were discussed: untyped `...any` checked at
runtime, one constructor per count with `any` arguments, and typed
declarations (`Key1[A]`…`Key3[A, B, C]`). Typed declarations won: a wrong
argument is a compile error, and the text is checked against the types when
it is declared (§5.8 of the interface).

## Decision

An error carries two messages. `Error()` is the code's text, for logs.
`T(ctx)` is for people and is picked from the error tree, not built by
wrapping: `Public`, `T` and `All` walk the tree in `errors.As` order, take
the first error that implements `i18n.Ter`, and never look inside it.
Declarations (`Key`, `Key1`–`Key3`) are package-level `i18n.Key` calls with
typed, checked arguments; an error argument composes (its public text fills
a slot) and is a cause; `Wrap` adds a cause people don't see. Kinds give an
HTTP status that follows the error people see. The full contract is
[interface.md](../interface.md) D1–D15 and §§4–5.

## Consequences

- Any wrapping works, and `i18n.Ter` stays the only contract, so packages
  that don't import this one can still make public errors (interface §6.5).
- Logs keep every cause: "log, then return a generic sentinel" becomes
  `Wrap`, logged once where the request ends.
- No generator and no registry: declaring is registering. A key used with
  two kinds or argument lists is refused by i18n's `Load` through meta
  (D12).
- An occurrence's `Error()` renders through i18n with no translator, so a
  deployment's edits never change logs.
- Composed messages are new: translators place the slot, and `%[1]s` lets
  them move it.
- Costs: the package shadows the standard library's name (it re-exports its
  API, so an app imports one `errors`); returning a `Def1`–`Def3` without
  its arguments isn't caught (a test that checks the message is); more than
  three arguments means splitting the message.
- i18n gains an amendment (interface §9) so labels passed as arguments
  render in the request's locale. This package prepares its own arguments
  and doesn't wait for it.
- zp's migration is a later, separate session (brief §11); it keeps its keys
  so its translations carry over.
