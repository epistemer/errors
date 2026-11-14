# errors

neozp's error package: errors with two messages, `Error()` for logs and
`T(ctx)` for people, through i18n.

- Read [docs/roadmap.md](docs/roadmap.md), then the newest entry of
  [docs/status.md](docs/status.md). The design is
  [ADR 0001](docs/adr/0001-two-messages.md), the locked contract
  [docs/interface.md](docs/interface.md); [docs/brief.md](docs/brief.md) is
  the history before them, with the acceptance tests (§7) and checked
  sketches (§8).
- Claude owns this repository until v1.0.0, on the terms of
  [docs/ownership.md](docs/ownership.md) (delegated 2026-11-14). API design
  is epistemer's; the interface was locked by epistemer on 2026-11-14.
- Load the `i18n` skill before touching anything that declares or renders a
  key.
- zp (`/home/epistemer/zp`) is private and read-only. Keep its names out of
  this repository, which is public: examples and tests use a toy shop.
  Counts and shapes are fine.
- Inside the package, import the standard library's errors as `stderrors`.
- Every working session adds an entry to docs/status.md.
