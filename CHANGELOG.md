# Changelog

All notable changes to this module. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and versions follow
[Semantic Versioning](https://semver.org/).

## [v0.1.0] - 2026-11-14

The first release: the whole of the interface locked on 2026-11-14.

### Added

- The standard library's API: `New`, `Is`, `As`, `AsType`, `Unwrap`, `Join`
  and `ErrUnsupported`, so an app imports one errors package.
- Declarations: `Key`, `Key1`, `Key2` and `Key3`, each an i18n key with
  meta of its kind and argument types. Texts are checked against their
  argument types when declared.
- Occurrences from `With` and `Wrap`: `Error()` for logs, `T(ctx)` for
  people, an error argument composed into the outer message, and `Is` true
  for the declaration.
- `Public`, `T` and `All`, which pick the message people see from the
  error tree; `Unknown` when nothing in it is public.
- `Kind`, eight kinds with HTTP statuses, and `KindOf`.

[v0.1.0]: https://github.com/epistemer/errors/releases/tag/v0.1.0
