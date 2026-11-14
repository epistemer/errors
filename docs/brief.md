# neozp `errors`: design brief

> Written on 2026-11-14 in a design session with epistemer, for the Claude Code session that starts `github.com/epistemer/errors`. epistemer accepted the design that day, including composed messages, which zp never supported. Decisions marked **(Claude)** were settled while writing this brief and can be overturned before the interface is locked. zp appears only as counts and shapes, so this file can be committed as `docs/brief.md`.

## 0. How to use this brief

This is neozp's fifth package, after cfgx, i18n, genx and models. The repository `github.com/epistemer/errors` exists and is empty.

1. Read the four sibling repositories for their conventions: the README, `docs/interface.md` (a locked contract with a decisions table and amendments), `docs/roadmap.md` (milestones, each item with a done-when), `docs/status.md`, `docs/adr/`, `docs/ownership.md`, `CLAUDE.md` and CI. Load the `i18n` skill.
2. Set the module up like theirs: Go 1.26, MIT, and the same CI jobs (Go 1.26 and 1.27 on Linux, macOS and Windows; gofmt, vet, staticcheck, tidy, tests with the race detector).
3. Write the design records from this brief: ADR 0001 from §§1–5, `docs/interface.md` from §§3–7 for epistemer to lock, and the roadmap from §12. Commit this file as `docs/brief.md`.
4. Build in the roadmap's order. §7 holds the acceptance tests; §8 has checked sketches of the subtle parts.
5. Ownership is epistemer's to grant when the work starts; the terms would be models' `docs/ownership.md`. Every working session adds an entry to `docs/status.md`.
6. zp is private and read-only. Keep its names out of this public repository: the examples use a toy shop like models' fixtures. Counts and shapes are fine.

## 1. The problem zp couldn't solve

zp has 454 `errors.New` sentinels, about half of them declared in templ files next to the forms that use them. A generator (`gen/error`) finds them in the AST and writes a registry that maps each message to an i18n key, deduplicated by message (about 400 remain). `errors.T(ctx, err)` looks up `err.Error()` in that registry.

`Error()` is the developers' message, and wrapping changes it, so a wrapped error rendered as "unknown error occurred". The code adapted by giving up wrapping:

- The service layer has no `fmt.Errorf` at all, and about 280 `return ErrUnknown` right after a `logrus.WithError(...)`: every site logs the cause and drops it.
- The database layer logs the driver's error and returns a sentinel in its place, so nothing above it can test for `context.Canceled`.
- API handlers return `err.Error()` (80 sites): clients get English whatever language they ask for, internals included.
- Every handler picks its HTTP status with its own `errors.Is` checks (37 in all).
- A stub that implemented both `error` and `i18n.Ter` failed from the other side: `fmt.Errorf("…: %w", e)` returns a `*fmt.wrapError`, which isn't a `Ter`.
- One message could never contain another, so a sentence like "checkout failed: insufficient balance" couldn't be translated as a whole.

## 2. The idea

An error carries two messages, and neither is derived from the other:

- **`Error()` is for logs.** It's the code's English text, it gains context with each wrap, and it may contain internals.
- **`T(ctx)` is for people.** It's in their locale, chosen by the code, and safe to show. It's **picked from the error tree, not built up by wrapping.**

So wrappers don't need to be translatable. `Public`, `T` and `All` walk the tree in `errors.As` order and take the first error that implements `i18n.Ter`. Every way of wrapping works (`%w`, two `%w`s, `errors.Join`, third-party wrappers) without knowing about i18n, and `i18n.Ter` stays the only contract.

A message that should contain another says so: its text has a slot, and the inner error is its argument. Translators decide where the slot goes.

## 3. Decisions

| # | Decision | Choice |
| :---- | :---- | :---- |
| D1 | What makes an error public | It implements `i18n.Ter`. No new interface, so packages that don't import this one can take part |
| D2 | Two messages | `Error()` is the code's text, for logs. `T(ctx)` is the text in ctx's locale, for people. People never see `Error()` text, not even through an argument |
| D3 | Finding the message | Walk the tree in `errors.As` order (pre-order, depth-first). The first public error wins, and the walk never looks inside a public error (§5.1) |
| D4 | Declaring | `Key`, `Key1[A]`, `Key2[A, B]` and `Key3[A, B, C]` in package-level variables. Each calls `i18n.Key`, so declaring is registering: no generator |
| D5 | Arguments | Typed by the declaration and checked when it's declared. A label renders in the same locale; an error renders as its public text and is a cause (§5.3, §5.8). **(Claude: typed arguments, chosen among the three fixes for argument counts that were discussed)** |
| D6 | Composition | An error argument puts its public text into the outer message, where translators place it (§5.4) |
| D7 | Causes people don't see | `Wrap`. `fmt.Errorf("%w: %w", d, cause)` still works, but makes the two siblings (§5.5) |
| D8 | Fallback | `Unknown`, a key like any other, so deployments can edit it. It also fills a slot whose error has no public part. **(Claude: the slot)** |
| D9 | Identity | An occurrence `Is` its declaration, through any wrapping (§5.6) |
| D10 | Kinds and status | Optional on a declaration. The status follows the error people see. A declaration without a kind takes the kind of its first error argument, else `Internal` (§5.7). **(Claude: that inheritance, and the eight kinds)** |
| D11 | Keys | Permanent: renaming one orphans its translations. New apps use `errors.<area>.<name>`. zp keeps its `errors.Err<Name>` keys, so its translations carry over |
| D12 | Meta | A declaration's kind and argument types are its i18n meta, so `Load` refuses one key with two meanings (§5.9). **(Claude)** |
| D13 | Package | `github.com/epistemer/errors`, package `errors`: a superset of the standard library's, so it replaces that import. Depends on i18n only. MIT |
| D14 | Style of texts | Like Go error strings: lowercase, no final period, so they compose in both messages. Capitalizing a message shown alone is the UI's job, e.g. CSS `::first-letter`. **(Claude)** |
| D15 | i18n | An amendment: i18n's `T` renders `Ter` arguments in the same context (§9). This package doesn't wait for it |

## 4. API

```go
package errors // import "github.com/epistemer/errors"

// ---- The standard library's API, so an app imports one errors package ----
// New makes a private error: people see Unknown in its place.

var ErrUnsupported = stderrors.ErrUnsupported // the standard library's

func New(text string) error
func Is(err, target error) bool
func As(err error, target any) bool
func AsType[E error](err error) (E, bool)
func Unwrap(err error) error
func Join(errs ...error) error

// ---- Declaring ----
// Only in package-level variable initializers: each calls i18n.Key, which
// panics after Load.

func Key(key, text string, opts ...Option) *Def
func Key1[A any](key, text string, opts ...Option) *Def1[A]
func Key2[A, B any](key, text string, opts ...Option) *Def2[A, B]
func Key3[A, B, C any](key, text string, opts ...Option) *Def3[A, B, C]

// Option configures a declaration. A Kind is an Option; v1 has no others.
type Option interface{ /* unexported method */ }

// Every declaration is an error and an i18n.TKer.

type Def struct{ /* unexported */ }

func (d *Def) Error() string                          // the code's text
func (d *Def) K() string                              // the key, which is also the error's code in an API
func (d *Def) T(ctx context.Context, _ ...any) string // the text in ctx's locale; args are ignored
func (d *Def) Kind() Kind                             // as declared; 0 when none
func (d *Def) Wrap(cause error) error                 // d over a cause people don't see; Wrap(nil) is d

type Def1[A any] struct{ /* unexported */ }

func (d *Def1[A]) Error() string                             // the code's text, unformatted
func (d *Def1[A]) K() string
func (d *Def1[A]) T(ctx context.Context, args ...any) string // prepares args as §5.3 says
func (d *Def1[A]) Kind() Kind
func (d *Def1[A]) With(a A) error                            // an occurrence
func (d *Def1[A]) Wrap(cause error, a A) error               // an occurrence over a cause people don't see

// Def2[A, B] and Def3[A, B, C] are the same with two and three arguments:
// With(a A, b B), Wrap(cause error, a A, b B), and so on.

// ---- Showing ----

// Unknown is what people see for an error with no public part.
var Unknown = Key("errors.unknown", "something went wrong", Internal)

func Public(err error) i18n.TKer             // the first public error in err's tree; Unknown if none; nil for nil
func T(ctx context.Context, err error) string // Public(err).T(ctx); "" for nil
func All(err error) iter.Seq[i18n.TKer]      // the outermost public error on each branch
func KindOf(err error) Kind                  // the kind of Public(err) (§5.7)

// ---- Kinds ----

type Kind uint8 // the zero Kind means none was declared

const (
	Internal        Kind = iota + 1 // 500
	Invalid                         // 400
	Unauthenticated                 // 401
	Forbidden                       // 403
	NotFound                        // 404
	Conflict                        // 409
	RateLimited                     // 429
	Unavailable                     // 503
)

func (k Kind) Status() int    // the HTTP status; 500 for the zero Kind
func (k Kind) String() string // "not_found"; "unspecified" for the zero Kind
```

An occurrence, which `With` and `Wrap` return, is unexported. It implements `error`, `i18n.TKer` and `Kind() Kind`, and has `Is(error) bool` and `Unwrap() []error`.

## 5. Rules

### 5.1 The walk

- An error in the tree is public when it implements `i18n.Ter` itself. Custom `As` methods aren't consulted.
- `Public(err)` is the first public error in pre-order, depth-first order (the order `errors.As` uses), `Unknown` when there is none, and nil for nil.
- `All(err)` yields public errors in the same order and never descends into one: the outermost public error on each branch. When it yields anything, its first value is `Public(err)`.
- `T(ctx, err)` is `Public(err).T(ctx)`, and `""` for nil.
- A public error that has no key (it implements `Ter` but not `Ker`) has the key `""`.

### 5.2 `Error()`

- A declaration's `Error()` is the code's text, unformatted.
- An occurrence's `Error()` is the code's text formatted with its arguments as logs see them (§5.3), then `": "` and the cause's `Error()` when `Wrap` gave one.
- It's never an edit or a translation: it renders through i18n with a context that holds no translator. A deployment's edits change what people see, never the logs.

### 5.3 Arguments

Each argument is prepared before formatting, in this order:

| Argument | For people (`T`) | For logs (`Error()`) |
| :---- | :---- | :---- |
| `nil` | `""` | as `fmt` prints it |
| an `error` | `T(ctx, arg)`: its public text, `Unknown`'s when it has none | `arg.Error()` |
| an `i18n.Ter` | `arg.T(ctx)` | `arg.T(context.Background())`, the code's text |
| anything else | unchanged | unchanged |

- Errors come before labels, because a public error is both.
- Error arguments are causes: `Unwrap` returns them (then `Wrap`'s cause), so `errors.Is` and `errors.As` see them.
- Numbers and dates go to `fmt` as they are. Locale-aware formatting is on i18n's roadmap.
- An occurrence's `T` ignores the `args` it's called with: `With` fixed its arguments. A `Def1`–`Def3` prepares the `args` its own `T` gets, so a declaration also works as a label.

### 5.4 Composition

- A declaration with an `error` parameter composes: its text has a slot for the inner error's public text, as in `Key1[error]("errors.checkout.failed", "checkout failed: %s")`.
- Translators place the slot; `%[1]s` lets them move it.
- It nests: the inner error can be composed itself, or wrapped by `fmt.Errorf` on the way.
- An inner error with no public part fills the slot with `Unknown`'s text.
- The inner error is a cause (§5.3): `errors.Is(err, ErrLowBalance)` holds for `ErrCheckoutFailed.With(ErrPaymentDeclined.With(ErrLowBalance))`. `Error()` composes the same way, with `Error()` texts.
- `All` yields a composed error once, because its inner messages are inside it.
- An error argument shows `Public`'s text, which is one message. To list several errors, use `All` in the UI.

### 5.5 Causes people don't see

- `d.Wrap(cause, …)` makes an occurrence whose `Unwrap` includes `cause`: the occurrence wraps the cause. People see `d`'s text; logs see `d`'s text, `": "` and `cause.Error()`.
- `fmt.Errorf("%w: %w", d, cause)` gives the same `Error()` and the same `Public`, but makes `d` and `cause` siblings under a `*fmt.wrapErrors`. If `cause` is public too, `All` yields both. Prefer `Wrap`.
- This replaces "log the error, then return ErrUnknown": wrap it, and log once where the request ends (§6.3).

### 5.6 Identity

- `errors.Is(err, d)` holds when `err`'s tree holds `d` or an occurrence of `d`: an occurrence's `Is` compares the target with its declaration.
- Two occurrences of one declaration aren't `Is` each other. Compare with the declaration.
- Occurrences are unexported, so code tests for a declaration with `Is`. `errors.AsType` works as usual for an app's own error types.

### 5.7 Kinds and status

- A declaration gets a kind as an option: `errors.Key("errors.shop.order_not_found", "order not found", errors.NotFound)`.
- `KindOf(err)` is the kind of `Public(err)`: its own when it declares one (a `Kind() Kind` method that returns non-zero); else, for an occurrence with error arguments, `KindOf` of its first error argument; else `Internal`. `KindOf(nil)` is `Internal`, so check `err` first.
- A cause given to `Wrap` never lends its kind. People don't see it, and the status follows what they see.
- A frame that composes other errors usually declares no kind, so "checkout failed: insufficient balance" is a 409 when `ErrLowBalance` is `Conflict`.
- Packages outside this one can give their public errors a kind by implementing `Kind() errors.Kind`.

| Kind | Status | For |
| :---- | :---- | :---- |
| none | 500 |  |
| `Internal` | 500 | the server's fault; `Unknown` |
| `Invalid` | 400 | the input is wrong |
| `Unauthenticated` | 401 | the caller isn't signed in |
| `Forbidden` | 403 | the caller may not do this |
| `NotFound` | 404 | no such thing |
| `Conflict` | 409 | the current state doesn't allow it |
| `RateLimited` | 429 | too many, too soon |
| `Unavailable` | 503 | try again later |

### 5.8 Checks when declaring

`Key1`–`Key3` check their text before they call `i18n.Key`, and panic with a message naming the key and the text unless:

1. `fmt` can format it with that many arguments: no `%!…(` in the output. That rules out missing, extra and bad-index arguments, and widths or precisions taken from arguments (`%*d`).
2. It uses every argument at least once.
3. Each verb suits its argument as `T` will pass it: errors and labels as strings, other concrete types as their zero value. Interface types other than `error` and `i18n.Ter` aren't checked, because their dynamic type isn't known.

`Key`'s text is never formatted, so `%` is an ordinary character in it. A text that needs arguments is declared with `Key1`–`Key3`; more than three means the message should be split.

What isn't caught: returning a `Def1`–`Def3` itself instead of an occurrence. Its text then shows its verbs. A test that checks the message catches it.

### 5.9 Meta

Each declaration passes `i18n.Key` a meta like `{"kind":"invalid","args":["i18n.Ter","int"]}`: the kind's `String()` and each argument type's `reflect.Type.String()`, with empty members left out, encoded by `encoding/json` from a struct so the bytes are stable. A declaration with neither has no meta, so it agrees with a plain `i18n.Key` of the same key and text.

- i18n's `Load` refuses a key registered with two metas, so nowhere in an app can one key have two kinds or two argument lists, and every such problem is listed at once.
- An editor built on i18n's `Search` can show translators what an error's arguments are.

## 6. Examples

A toy shop. Labels such as `L10nAmount` are ordinary `i18n.Key` handles, as cfgx or models generate them.

### 6.1 Declaring

```go
var (
	ErrOrderNotFound = errors.Key("errors.shop.order_not_found", "order not found", errors.NotFound)
	ErrOrderNotSaved = errors.Key("errors.shop.order_not_saved", "the order couldn't be saved")
	ErrLowBalance    = errors.Key("errors.wallet.low_balance", "insufficient balance", errors.Conflict)

	ErrRequired = errors.Key1[i18n.Ter]("errors.form.required", "%s is required", errors.Invalid)
	ErrTooHigh  = errors.Key2[i18n.Ter, int]("errors.form.too_high", "%s must be at most %d", errors.Invalid)

	// Frames around another error's message: no kind, so they take the inner one's.
	ErrPaymentDeclined = errors.Key1[error]("errors.payment.declined", "payment declined: %s")
	ErrCheckoutFailed  = errors.Key1[error]("errors.checkout.failed", "checkout failed: %s")
)
```

### 6.2 Returning

```go
func (s *Shop) Checkout(ctx context.Context, id int64) error {
	o, err := s.q.Order(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrOrderNotFound // a declaration is its own sentinel
	}
	if err != nil {
		return fmt.Errorf("load order %d: %w", id, err) // context for logs only
	}
	if err := s.wallet.Debit(ctx, o.CustomerID, o.Total); err != nil {
		return ErrCheckoutFailed.With(err) // composed: the wallet's message goes inside
	}
	if err := s.q.MarkPaid(ctx, id); err != nil {
		return ErrOrderNotSaved.Wrap(err) // people see ours; logs and errors.Is keep the driver's
	}
	return nil
}

func (w *Wallet) Debit(ctx context.Context, customer, amount int64) error {
	// …
	if balance < amount {
		return ErrPaymentDeclined.With(ErrLowBalance)
	}
	// …
}

func validate(name string, amount int) error {
	var errs []error
	if name == "" {
		errs = append(errs, ErrRequired.With(L10nName))
	}
	if amount > 5000 {
		errs = append(errs, ErrTooHigh.With(L10nAmount, 5000))
	}
	return errors.Join(errs...)
}
```

### 6.3 Showing

In templ, as zp writes it today:

```
templ InputErr(err error) {
	<p class="first-letter:uppercase text-sm text-red-600">{ errors.T(ctx, err) }</p>
}

templ FormErrors(err error) {
	<ul>
		for p := range errors.All(err) {
			<li class="first-letter:uppercase">{ p.T(ctx) }</li>
		}
	</ul>
}
```

In an API, where a request ends:

```go
// writeError ends a failed API request. The status, the code and the message
// all follow the error people see, and only server errors are logged.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	ctx := r.Context()
	p, kind := errors.Public(err), errors.KindOf(err)
	if kind.Status() >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, "request failed", "err", err, "code", p.K())
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(kind.Status())
	json.NewEncoder(w).Encode(struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{p.K(), p.T(ctx)})
}
```

### 6.4 What logs and people see

With the Arabic translations of §7:

| The error | `Error()` | People, in Arabic | Code | Status |
| :---- | :---- | :---- | :---- | :---- |
| `fmt.Errorf("load order 42: %w", ErrOrderNotFound)` | load order 42: order not found | الطلب غير موجود | `errors.shop.order_not_found` | 404 |
| `ErrOrderNotSaved.Wrap(context.DeadlineExceeded)` | the order couldn't be saved: context deadline exceeded | تعذر حفظ الطلب | `errors.shop.order_not_saved` | 500 |
| `fmt.Errorf("query orders: %w", sql.ErrConnDone)` | query orders: sql: connection is already closed | حدث خطأ ما | `errors.unknown` | 500 |
| `ErrTooHigh.With(L10nAmount, 5000)` | Amount must be at most 5000 | يجب ألا يتجاوز المبلغ 5000 | `errors.form.too_high` | 400 |
| `ErrCheckoutFailed.With(ErrPaymentDeclined.With(ErrLowBalance))` | checkout failed: payment declined: insufficient balance | تعذر إتمام الطلب: رُفض الدفع: الرصيد غير كافٍ | `errors.checkout.failed` | 409 |
| `ErrPaymentDeclined.With(sql.ErrConnDone)` | payment declined: sql: connection is already closed | رُفض الدفع: حدث خطأ ما | `errors.payment.declined` | 500 |

### 6.5 A package that doesn't use declarations

Any error type that implements `i18n.Ter` is public. `Kind()` is optional:

```go
type quotaError struct{ limit int }

func (e quotaError) Error() string { return fmt.Sprintf("quota of %d exceeded", e.limit) }
func (e quotaError) T(ctx context.Context, _ ...any) string {
	return ErrTooHigh.T(ctx, L10nAmount, e.limit)
}
func (quotaError) Kind() errors.Kind { return errors.RateLimited }
```

## 7. Acceptance tests

These were checked with a stand-in for i18n that has i18n's API shapes and resolution order (the locale's edit, then the source locale's, then the code's text): 21 cases in 6 tests pass with `-race`. Make them pass against the real i18n through `i18ntest`.

```go
package errors_test

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	"github.com/epistemer/i18n/i18ntest"
)

// Labels, as models or cfgx would generate them.
var (
	L10nAmount = i18n.Key("form.amount", "Amount")
	L10nName   = i18n.Key("form.name", "Name")
)

var (
	ErrOrderNotFound = errors.Key("errors.shop.order_not_found", "order not found", errors.NotFound)
	ErrOrderNotSaved = errors.Key("errors.shop.order_not_saved", "the order couldn't be saved")
	ErrLowBalance    = errors.Key("errors.wallet.low_balance", "insufficient balance", errors.Conflict)

	ErrRequired = errors.Key1[i18n.Ter]("errors.form.required", "%s is required", errors.Invalid)
	ErrTooHigh  = errors.Key2[i18n.Ter, int]("errors.form.too_high", "%s must be at most %d", errors.Invalid)

	// Frames around another error's message. They declare no kind, so
	// KindOf takes the inner error's.
	ErrPaymentDeclined = errors.Key1[error]("errors.payment.declined", "payment declined: %s")
	ErrCheckoutFailed  = errors.Key1[error]("errors.checkout.failed", "checkout failed: %s")
	ErrImportLine      = errors.Key2[int, error]("errors.import.line", "line %d: %s")
)

var arabic = map[string]string{
	"errors.shop.order_not_found": "الطلب غير موجود",
	"errors.shop.order_not_saved": "تعذر حفظ الطلب",
	"errors.wallet.low_balance":   "الرصيد غير كافٍ",
	"errors.form.required":        "%s مطلوب",
	"errors.form.too_high":        "يجب ألا يتجاوز %[1]s %[2]d", // reordered: allowed
	"errors.payment.declined":     "رُفض الدفع: %s",
	"errors.checkout.failed":      "تعذر إتمام الطلب: %s",
	"errors.import.line":          "السطر %d: %s",
	"errors.unknown":              "حدث خطأ ما",
	"form.amount":                 "المبلغ",
	"form.name":                   "الاسم",
}

// contexts returns an Arabic request, and an English one on a deployment
// that edited one English text (a white-label edit).
func contexts(t *testing.T) (ar, en context.Context) {
	t.Helper()
	tr := i18ntest.New(t, i18n.Locale{Code: "en", NativeName: "English"})
	if err := tr.CreateLocale(i18n.Locale{Code: "ar", NativeName: "العربية", RTL: true}); err != nil {
		t.Fatal(err)
	}
	for key, text := range arabic {
		if err := tr.Set("ar", key, text); err != nil {
			t.Fatal(err)
		}
	}
	if err := tr.Set("en", "errors.shop.order_not_found", "We couldn't find that order"); err != nil {
		t.Fatal(err)
	}
	return i18ntest.Context(t, tr, "ar"), i18ntest.Context(t, tr, "en")
}

// quotaError is public without this package's declarations: it implements
// i18n.Ter, and Kind() for its status. It has no key.
type quotaError struct{ limit int }

func (e quotaError) Error() string { return fmt.Sprintf("quota of %d exceeded", e.limit) }
func (e quotaError) T(ctx context.Context, _ ...any) string {
	return ErrTooHigh.T(ctx, L10nAmount, e.limit)
}
func (quotaError) Kind() errors.Kind { return errors.RateLimited }

func TestShow(t *testing.T) {
	ar, _ := contexts(t)
	tests := []struct {
		err             error
		log, user, code string
		kind            errors.Kind
	}{{
		err:  fmt.Errorf("load order 42: %w", ErrOrderNotFound),
		log:  "load order 42: order not found",
		user: "الطلب غير موجود", code: "errors.shop.order_not_found", kind: errors.NotFound,
	}, {
		err:  ErrOrderNotSaved.Wrap(context.DeadlineExceeded),
		log:  "the order couldn't be saved: context deadline exceeded",
		user: "تعذر حفظ الطلب", code: "errors.shop.order_not_saved", kind: errors.Internal,
	}, {
		err:  fmt.Errorf("query orders: %w", sql.ErrConnDone),
		log:  "query orders: sql: connection is already closed",
		user: "حدث خطأ ما", code: "errors.unknown", kind: errors.Internal,
	}, {
		err:  ErrTooHigh.With(L10nAmount, 5000),
		log:  "Amount must be at most 5000",
		user: "يجب ألا يتجاوز المبلغ 5000", code: "errors.form.too_high", kind: errors.Invalid,
	}, {
		err:  ErrCheckoutFailed.With(ErrPaymentDeclined.With(ErrLowBalance)),
		log:  "checkout failed: payment declined: insufficient balance",
		user: "تعذر إتمام الطلب: رُفض الدفع: الرصيد غير كافٍ", code: "errors.checkout.failed", kind: errors.Conflict,
	}, {
		err:  ErrPaymentDeclined.With(fmt.Errorf("debit wallet 7: %w", ErrLowBalance)),
		log:  "payment declined: debit wallet 7: insufficient balance",
		user: "رُفض الدفع: الرصيد غير كافٍ", code: "errors.payment.declined", kind: errors.Conflict,
	}, {
		err:  ErrPaymentDeclined.With(sql.ErrConnDone),
		log:  "payment declined: sql: connection is already closed",
		user: "رُفض الدفع: حدث خطأ ما", code: "errors.payment.declined", kind: errors.Internal,
	}, {
		err:  ErrImportLine.With(7, ErrTooHigh.With(L10nAmount, 5000)),
		log:  "line 7: Amount must be at most 5000",
		user: "السطر 7: يجب ألا يتجاوز المبلغ 5000", code: "errors.import.line", kind: errors.Invalid,
	}, {
		err:  ErrOrderNotSaved.Wrap(ErrOrderNotFound),
		log:  "the order couldn't be saved: order not found",
		user: "تعذر حفظ الطلب", code: "errors.shop.order_not_saved", kind: errors.Internal,
	}, {
		err:  fmt.Errorf("transfer: %w", quotaError{limit: 300}),
		log:  "transfer: quota of 300 exceeded",
		user: "يجب ألا يتجاوز المبلغ 300", code: "", kind: errors.RateLimited,
	}}
	for _, tt := range tests {
		t.Run(tt.log, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.log {
				t.Errorf("Error() = %q, want %q", got, tt.log)
			}
			if got := errors.T(ar, tt.err); got != tt.user {
				t.Errorf("T = %q, want %q", got, tt.user)
			}
			if got := errors.Public(tt.err).K(); got != tt.code {
				t.Errorf("Public().K() = %q, want %q", got, tt.code)
			}
			if got := errors.KindOf(tt.err); got != tt.kind {
				t.Errorf("KindOf = %v, want %v", got, tt.kind)
			}
		})
	}
}

func TestIdentity(t *testing.T) {
	tests := []struct {
		err, target error
		want        bool
	}{
		{fmt.Errorf("load: %w", ErrOrderNotFound), ErrOrderNotFound, true},
		{ErrTooHigh.With(L10nAmount, 1), ErrTooHigh, true},
		{ErrTooHigh.With(L10nAmount, 1), ErrRequired, false},
		{ErrCheckoutFailed.With(ErrPaymentDeclined.With(ErrLowBalance)), ErrLowBalance, true},
		{ErrOrderNotSaved.Wrap(context.DeadlineExceeded), context.DeadlineExceeded, true},
		{ErrOrderNotSaved.Wrap(context.DeadlineExceeded), ErrOrderNotSaved, true},
		{errors.Join(ErrRequired.With(L10nName), ErrOrderNotFound), ErrOrderNotFound, true},
	}
	for _, tt := range tests {
		if got := errors.Is(tt.err, tt.target); got != tt.want {
			t.Errorf("Is(%q, %q) = %v, want %v", tt.err, tt.target, got, tt.want)
		}
	}
}

func messages(ctx context.Context, err error) []string {
	var out []string
	for p := range errors.All(err) {
		out = append(out, p.T(ctx))
	}
	return out
}

func TestAll(t *testing.T) {
	ar, _ := contexts(t)
	tests := []struct {
		name string
		err  error
		want []string
	}{{
		name: "a form: one message per branch",
		err: fmt.Errorf("validate order: %w", errors.Join(
			ErrRequired.With(L10nName), ErrTooHigh.With(L10nAmount, 5000))),
		want: []string{"الاسم مطلوب", "يجب ألا يتجاوز المبلغ 5000"},
	}, {
		name: "a composed error is one message",
		err:  ErrCheckoutFailed.With(ErrPaymentDeclined.With(ErrLowBalance)),
		want: []string{"تعذر إتمام الطلب: رُفض الدفع: الرصيد غير كافٍ"},
	}, {
		name: "Wrap hides a public cause",
		err:  ErrOrderNotSaved.Wrap(ErrOrderNotFound),
		want: []string{"تعذر حفظ الطلب"},
	}, {
		name: "two %w make siblings",
		err:  fmt.Errorf("%w: %w", ErrOrderNotSaved, ErrOrderNotFound),
		want: []string{"تعذر حفظ الطلب", "الطلب غير موجود"},
	}, {
		name: "nothing public",
		err:  fmt.Errorf("query: %w", sql.ErrConnDone),
		want: nil,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := messages(ar, tt.err)
			if strings.Join(got, "|") != strings.Join(tt.want, "|") {
				t.Errorf("All = %q, want %q", got, tt.want)
			}
			if len(got) > 0 && errors.T(ar, tt.err) != got[0] {
				t.Errorf("T = %q, want the first of All, %q", errors.T(ar, tt.err), got[0])
			}
		})
	}
}

func TestWhiteLabel(t *testing.T) {
	_, en := contexts(t)
	err := fmt.Errorf("load order 42: %w", ErrOrderNotFound)
	if got, want := errors.T(en, err), "We couldn't find that order"; got != want {
		t.Errorf("T = %q, want the deployment's edit %q", got, want)
	}
	if got, want := err.Error(), "load order 42: order not found"; got != want {
		t.Errorf("Error() = %q, want the code's text %q", got, want)
	}
}

func TestNil(t *testing.T) {
	if errors.Public(nil) != nil || errors.T(context.Background(), nil) != "" {
		t.Error("Public(nil) must be nil and T(ctx, nil) empty")
	}
}

// People never see Error() text: random trees of wrapped, joined and
// composed errors, whose private texts all carry a marker.
var (
	propFrame = errors.Key1[error]("errors.test.frame", "frame: %s")
	propPair  = errors.Key2[int, error]("errors.test.pair", "%d: %s")
	propPlain = errors.Key("errors.test.plain", "plain")
)

func TestNoPrivateTextReachesPeople(t *testing.T) {
	ar, _ := contexts(t)
	const secret = "SECRET"
	r := rand.New(rand.NewPCG(1, 2))
	var gen func(depth int) error
	gen = func(depth int) error {
		if depth == 0 {
			switch r.IntN(3) {
			case 0:
				return fmt.Errorf("db %s %d", secret, r.Int())
			case 1:
				return fmt.Errorf("%s: %w", secret, sql.ErrConnDone)
			default:
				return propPlain
			}
		}
		switch r.IntN(6) {
		case 0:
			return fmt.Errorf("ctx %s: %w", secret, gen(depth-1))
		case 1:
			return fmt.Errorf("%w: %w", gen(depth-1), gen(depth-1))
		case 2:
			return errors.Join(gen(depth-1), gen(depth-1))
		case 3:
			return propFrame.With(gen(depth - 1))
		case 4:
			return propPair.Wrap(gen(depth-1), r.IntN(9), gen(depth-1))
		default:
			return propPlain.Wrap(gen(depth - 1))
		}
	}
	for i := range 2000 {
		err := gen(1 + i%5)
		for _, s := range append(messages(ar, err), errors.T(ar, err)) {
			if strings.Contains(s, secret) {
				t.Fatalf("people see private text %q\nError(): %s", s, err)
			}
		}
	}
}
```

The declaration checks (§5.8) are tested on the checker itself, in an internal test: a declaration that panics can't be a package-level variable, and one that passes can't be made in a test after `Load`. These cases were checked:

| Text | Argument types | Result |
| :---- | :---- | :---- |
| `%s is required` | `i18n.Ter` | accepted |
| `100%% of %s` | `i18n.Ter` | accepted |
| `%[2]d for %[1]s` | `i18n.Ter`, `int` | accepted |
| `%v` | `any` | accepted |
| `%d` | `any` | accepted: the dynamic type isn't known |
| `line %d: %s` | `int`, `error` | accepted |
| `no verbs` | `i18n.Ter` | panics: argument 1 unused |
| `%s only` | `i18n.Ter`, `int` | panics: argument 2 unused |
| `%s and %s` | `i18n.Ter` | panics: an argument is missing |
| `%s` | `int` | panics: the verb doesn't suit `int` |
| `%d` | `error` | panics: errors render as text |
| `%*d` | `int` | panics: a width taken from an argument |
| `%[3]s` | `i18n.Ter` | panics: bad index |
| `50%` | `i18n.Ter` | panics: no verb after `%` |

## 8. Implementation notes

- Inside the package, import the standard library's package as `stderrors`.
- The declarations share an unexported core: the i18n handle and the kind. Write each type's methods out, so godoc lists them on every type.
- An occurrence holds its declaration (as an `error`, for `Is`), the arguments, and `Wrap`'s cause. Its unexported `shownErrors()` returns its error arguments without the cause; `KindOf` uses it (§5.7).
- `Key1`–`Key3` get their argument types with `reflect.TypeFor[A]()`, run the checks, then call `i18n.Key(key, text, metaOf(kind, types)...)`.
- Nothing is cached and nothing is mutable after package initialization, so every function is safe for concurrent use.

These sketches pass the tests of §7 against the stand-in; names are free.

The walk, and what it serves:

```go
// Public returns the first error in err's tree, in errors.As order, that
// implements i18n.Ter; Unknown when there is none; nil for nil.
func Public(err error) i18n.TKer {
	if err == nil {
		return nil
	}
	if n := first(err); n != nil {
		return asTKer(n)
	}
	return Unknown
}

// T is Public(err).T(ctx), or "" for nil.
func T(ctx context.Context, err error) string {
	if err == nil {
		return ""
	}
	return Public(err).T(ctx)
}

// All yields the outermost public error on each branch of err's tree.
func All(err error) iter.Seq[i18n.TKer] {
	return func(yield func(i18n.TKer) bool) {
		walk(err, func(n error) bool { return yield(asTKer(n)) })
	}
}

// KindOf is the kind of Public(err): its own, else that of its first error
// argument, else Internal.
func KindOf(err error) Kind {
	n := first(err)
	if n == nil {
		return Internal
	}
	if k, ok := n.(interface{ Kind() Kind }); ok && k.Kind() != 0 {
		return k.Kind()
	}
	if o, ok := n.(interface{ shownErrors() []error }); ok {
		if errs := o.shownErrors(); len(errs) > 0 {
			return KindOf(errs[0])
		}
	}
	return Internal
}

func first(err error) (found error) {
	walk(err, func(n error) bool { found = n; return false })
	return found
}

// walk calls f on each public node that no other public node wraps, in
// pre-order; it doesn't descend into a public node.
func walk(err error, f func(error) bool) bool {
	if err == nil {
		return true
	}
	if _, ok := err.(i18n.Ter); ok {
		return f(err)
	}
	switch u := err.(type) {
	case interface{ Unwrap() error }:
		return walk(u.Unwrap(), f)
	case interface{ Unwrap() []error }:
		for _, e := range u.Unwrap() {
			if !walk(e, f) {
				return false
			}
		}
	}
	return true
}

type noKey struct{ i18n.Ter }

func (noKey) K() string { return "" }

func asTKer(n error) i18n.TKer {
	if tk, ok := n.(i18n.TKer); ok {
		return tk
	}
	return noKey{n.(i18n.Ter)}
}
```

Arguments (§5.3):

```go
// forPeople prepares arguments for people: an error shows as its public
// text, a label in the same locale. An error's Error() never reaches them.
func forPeople(ctx context.Context, args []any) []any {
	if len(args) == 0 {
		return nil
	}
	out := make([]any, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case nil:
			out[i] = ""
		case error:
			out[i] = T(ctx, v)
		case i18n.Ter:
			out[i] = v.T(ctx)
		default:
			out[i] = a
		}
	}
	return out
}

// forLogs prepares arguments for Error(): Error(), and the code's text.
func forLogs(args []any) []any {
	if len(args) == 0 {
		return nil
	}
	out := make([]any, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case error:
			out[i] = v.Error()
		case i18n.Ter:
			out[i] = v.T(context.Background())
		default:
			out[i] = a
		}
	}
	return out
}
```

The checks (§5.8). `fmt` is the parser, so there's no copy of i18n's:

```go
var (
	terType   = reflect.TypeFor[i18n.Ter]()
	errorType = reflect.TypeFor[error]()
	badFmt    = regexp.MustCompile(`%!.?\(`)
)

type probe int

func (p probe) Format(f fmt.State, _ rune) { fmt.Fprintf(f, "\x00%d\x00", int(p)) }

// checkText panics unless text uses each argument, and fmt can format it
// with values of the argument types as T will see them.
func checkText(key, text string, ts []reflect.Type) {
	probes := make([]any, len(ts))
	for i := range ts {
		probes[i] = probe(i)
	}
	out := fmt.Sprintf(text, probes...)
	if badFmt.MatchString(out) {
		panic(fmt.Sprintf("errors: %s: fmt can't format %q with %d arguments", key, text, len(ts)))
	}
	for i := range ts {
		if !strings.Contains(out, fmt.Sprintf("\x00%d\x00", i)) {
			panic(fmt.Sprintf("errors: %s: %q doesn't use argument %d", key, text, i+1))
		}
	}
	zeros := make([]any, len(ts))
	for i, t := range ts {
		switch {
		case t.Implements(errorType), t.Implements(terType):
			zeros[i] = "" // T turns these into text
		case t.Kind() == reflect.Interface:
			zeros[i] = probe(i) // the dynamic type isn't known
		default:
			zeros[i] = reflect.Zero(t).Interface()
		}
	}
	if out := fmt.Sprintf(text, zeros...); badFmt.MatchString(out) {
		panic(fmt.Sprintf("errors: %s: a verb in %q doesn't suit its argument's type: %s", key, text, out))
	}
}
```

Meta (§5.9):

```go
type meta struct {
	Kind string   `json:"kind,omitempty"`
	Args []string `json:"args,omitempty"`
}

// metaOf is a declaration's i18n meta: nil without a kind and arguments.
func metaOf(k Kind, ts []reflect.Type) []byte {
	var m meta
	if k != 0 {
		m.Kind = k.String()
	}
	for _, t := range ts {
		m.Args = append(m.Args, t.String())
	}
	if m.Kind == "" && len(m.Args) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err) // strings only: can't happen
	}
	return b
}
```

## 9. The i18n amendment

i18n's interface §5.2 gains one rule, recorded in its §13:

> An argument that implements `Ter` is rendered with the same context before the text is formatted, so `Hello.T(ctx, settings.L10nServer)` shows the label in the request's locale. With no translator in the context, that's the label's code text.
>

- Today such an argument goes to `fmt` as it is, which prints a struct dump, because the key type has no `String` method.
- Render the arguments before `T` opens its read transaction, in `resolveCtx` ahead of `db.View`. bbolt warns that even a nested read-only transaction in one goroutine can deadlock: the inner one can wait for a writer that is waiting for the outer one.
- The fit rules don't change: the argument is a string by the time the text is formatted, so the code's text uses `%s` or `%v` for it.
- It's made in the i18n repository, under i18n's terms, and released as `v0.3.0`. This package prepares its own arguments, so it doesn't wait for it.
- **Done when:** tests show a label argument in the request's locale, and in the code's text without a translator; a label whose own text takes a label renders both in the locale; a stress test with concurrent `Set`s runs ten minutes clean under `-race`; the fit tests are unchanged.

## 10. Out of v1

- Field paths on errors. Forms keep one error per field, or list them with `All`.
- Stack traces.
- An HTTP writer in the package. Apps own their response bodies; §6.3 shows one.
- The codes of composed inner errors in an API body, which carries the outer code only.
- A different text for `Unknown` inside a slot than alone.
- More than one message in a slot: an error argument shows `Public`'s.
- Plural forms and locale-aware numbers in arguments, which are on i18n's roadmap.
- A vet check for a `Def1`–`Def3` returned without its arguments.

## 11. zp: what changes

For zp's own session, later. Nothing in this repository touches zp.

- Each of the 454 declarations becomes `errors.Key("errors.Err<Name>", "<message>")`, with the key zp's generated registry gives its message. The registry deduplicated by message, and the variable whose name sorts first gave the key, so variables that share a message share a key, and the translations in zp's store carry over. Variables that share a key must agree on its kind (§5.9).
- Kinds go where handlers map statuses today (the 37 `errors.Is` checks).
- `gen/error` and the generated registry go away. neozp's i18n already has no error registry.
- The ~280 sites that log and return `ErrUnknown` return `Wrap` or `fmt.Errorf` instead, and the request's end logs once.
- The database layer keeps the driver's error as a cause.
- The 80 `err.Error()` in API handlers become a helper like §6.3's.
- The five `errors.T(ctx, err)` call sites (alerts, input errors, the 404 and 500 pages, flash messages) keep their shape.
- zp's `errors.Unknown` key ("unknown error occurred") gives way to `errors.unknown`. i18n can't rename keys yet, so its translations are entered again.

## 12. Roadmap to v0.1.0

Each item lands as a PR with green CI and a review by the other model family. A milestone is done when every item meets its done-when.

### M0. Foundations

- [ ] **M0.1** Module `github.com/epistemer/errors` (Go 1.26, MIT), `doc.go`, `CLAUDE.md`, and CI like i18n's. *Done when:* CI runs green on a PR.
- [ ] **M0.2** ADR 0001 (§§1–5), `docs/interface.md` (§§3–7), this roadmap, `docs/status.md`, and this file as `docs/brief.md`. *Done when:* epistemer locks the interface.

### M1. Declaring

- [ ] **M1.1** The standard library's API (§4). *Done when:* a test shows each one behaves as the standard library's, `AsType` included.
- [ ] **M1.2** `Kind` and `Option`: names, statuses, `String`. *Done when:* a table test covers every kind and the zero Kind.
- [ ] **M1.3** `Key`–`Key3`, the checks (§5.8) and meta (§5.9). *Done when:* the checker's internal test covers §7's fourteen cases; meta is byte-stable and absent for a plain `Key`; a test package of its own shows `Load` refusing one key declared with two kinds.
- [ ] **M1.4** Occurrences: `With`, `Wrap`, `Is`, `Unwrap`, `Error()`. *Done when:* `TestIdentity` passes, and `Error()` stays the code's text with an Arabic edit and an English edit in the store.

### M2. Showing

- [ ] **M2.1** The walk: `Public`, `T`, `All` (§5.1). *Done when:* `TestAll` and `TestNil` pass.
- [ ] **M2.2** Arguments and composition (§5.3, §5.4). *Done when:* `TestShow`, `TestWhiteLabel` and `TestNoPrivateTextReachesPeople` pass.
- [ ] **M2.3** `KindOf` (§5.7). *Done when:* `TestShow`'s kinds pass, including inheritance through two frames and a `Wrap` cause that lends nothing.

### M3. Release

- [ ] **M3.1** Godoc examples, and a README with a quickstart. *Done when:* the examples run as tests, and the quickstart's output is pasted from a real run.
- [ ] **M3.2** API review of every exported identifier against the interface. *Done when:* a written review lists each one and how it was checked, and its findings are fixed or filed.
- [ ] **M3.3** CHANGELOG, the `v0.1.0` tag, and a global Claude Code skill (`~/.claude/skills/errors/`) like cfgx's, i18n's and genx's. *Done when:* `go get github.com/epistemer/errors@v0.1.0` works through the module proxy.

### Alongside

- [ ] **i18n amendment** (§9), in the i18n repository.

### Working with agents

As in the sibling repositories: Claude owns the cross-cutting core (M1.3, M1.4, M2.1–M2.3) and reviews Gemini's work; Gemini, through agy, takes leaf items with a crisp done-when (M0.1, M1.1, M1.2, M3.1). Each delegated item runs in its own git worktree outside the repository.
