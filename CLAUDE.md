# errors

neozp's error package: errors with two messages, `Error()` for logs and
`T(ctx)` for people, through i18n.

- Read [docs/roadmap.md](docs/roadmap.md), then the newest entry of
  [docs/status.md](docs/status.md). The design is
  [ADR 0001](docs/adr/0001-two-messages.md), the contract
  [docs/interface.md](docs/interface.md); [docs/brief.md](docs/brief.md) is
  the history before them, with the acceptance tests (§7) and checked
  sketches (§8).
- Ownership is epistemer's to grant ([docs/ownership.md](docs/ownership.md)
  holds the proposed terms, models'). Until it's granted, bring each
  milestone to epistemer. API design and the interface lock are epistemer's
  either way.
- Load the `i18n` skill before touching anything that declares or renders a
  key.
- zp (`/home/epistemer/zp`) is private and read-only. Keep its names out of
  this repository, which is public: examples and tests use a toy shop.
  Counts and shapes are fine.
- Inside the package, import the standard library's errors as `stderrors`.
- Every working session adds an entry to docs/status.md.
