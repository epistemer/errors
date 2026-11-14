# errors

neozp's error package. An error carries two messages, and neither is derived
from the other:

- `Error()` is for logs: the code's English text, gaining context with each
  wrap.
- `T(ctx)` is for people: text in their locale, through
  [i18n](https://github.com/epistemer/i18n), picked from the error tree
  rather than built up by wrapping, so `%w`, `errors.Join` and third-party
  wrappers all keep working.

A drop-in superset of the standard library's `errors`, with typed, composable
declarations (`errors.Key1[error]("errors.checkout.failed", "checkout failed: %s")`)
and an HTTP status per kind.

**In development.** Nothing to use yet:

- [ADR 0001](docs/adr/0001-two-messages.md): the design and why.
- [docs/interface.md](docs/interface.md): the contract, locked 2026-11-14.
- [docs/roadmap.md](docs/roadmap.md): the plan to `v0.1.0`, with a
  done-when for each item.
- [docs/status.md](docs/status.md): the log of working sessions.
- [docs/brief.md](docs/brief.md): the design brief the work started from.

Part of neozp, alongside [cfgx](https://github.com/epistemer/cfgx),
[i18n](https://github.com/epistemer/i18n),
[genx](https://github.com/epistemer/genx) and
[models](https://github.com/epistemer/models). License: [MIT](LICENSE).
