# errors — Developer Interface (v1, draft)

> Status: **draft, waiting for epistemer to lock it.** Written on 2026-11-14
> from the design brief ([brief.md](brief.md) §§3–7, unchanged in substance),
> which epistemer accepted that day. Once locked, any change needs an
> explicit amendment, recorded in §10. This document covers what a developer
> writes and what they get back. Why it is shaped this way is
> [ADR 0001](adr/0001-two-messages.md); the order of work is
> [roadmap.md](roadmap.md).
>
> Decisions marked **(Claude)** were settled while the brief was written and
> can still be overturned before the lock.
>
> Section numbers follow the brief's, so its cross-references hold.

## 1. Scope

`github.com/epistemer/errors`, package `errors`, replaces the standard
library's `errors` import in neozp apps. It depends on
`github.com/epistemer/i18n` only.

## 2. Principles

1. **Two messages, neither derived from the other.** `Error()` is the code's
   English text for logs, and gains context with each wrap. `T(ctx)` is the
   text in the request's locale, for people.
2. **The message people see is picked from the tree, not built by
   wrapping.** Any wrapper works without knowing about i18n; `i18n.Ter` is
   the only contract.
3. **People never see `Error()` text,** not even through an argument.
4. **Declaring is registering.** No generator: a declaration is an
   `i18n.Key` call in a package-level variable.
5. **Composition is explicit.** A message that contains another has a slot,
   and translators place it.

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

// Every declaration is an error and an i18n.Ker. Def is also an i18n.TKer.
// Def1–Def3 take their arguments in a typed T and aren't i18n.Ters, so one
// returned without its arguments isn't public: people see Unknown (§5.8).

type Def struct{ /* unexported */ }

func (d *Def) Error() string                          // the code's text
func (d *Def) K() string                              // the key, which is also the error's code in an API
func (d *Def) T(ctx context.Context, _ ...any) string // the text in ctx's locale; args are ignored
func (d *Def) Kind() Kind                             // as declared; 0 when none
func (d *Def) Wrap(cause error) error                 // d over a cause people don't see; Wrap(nil) is d

type Def1[A any] struct{ /* unexported */ }

func (d *Def1[A]) Error() string                             // the code's text, unformatted
func (d *Def1[A]) K() string
func (d *Def1[A]) T(ctx context.Context, a A) string          // the text in ctx's locale, a prepared as §5.3 says
func (d *Def1[A]) Kind() Kind
func (d *Def1[A]) With(a A) error                            // an occurrence
func (d *Def1[A]) Wrap(cause error, a A) error               // an occurrence over a cause people don't see

// Def2[A, B] and Def3[A, B, C] are the same with two and three arguments:
// T(ctx, a A, b B), With(a A, b B), Wrap(cause error, a A, b B), and so on.

// ---- Showing ----

// Unknown is what people see for an error with no public part.
var Unknown = Key("errors.unknown", "something went wrong", Internal)

func Public(err error) i18n.TKer             // the first public error in err's tree; Unknown if none; nil for nil
func T(ctx context.Context, err error) string // Public(err).T(ctx); "" for nil
func All(err error) iter.Seq[i18n.TKer]      // the outermost public error on each branch
func KindOf(err error) Kind                  // the kind of Public(err) (§5.7); Internal for nil

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
- A nil pointer in the tree, or given as a cause or an argument, counts as nil.
- Like `errors.Is`, the walk assumes the tree has no cycles.

### 5.2 `Error()`

- A declaration's `Error()` is the code's text, unformatted.
- An occurrence's `Error()` is the code's text formatted with its arguments as logs see them (§5.3), then `": "` and the cause's `Error()` when `Wrap` gave one.
- It's never an edit or a translation: it renders through i18n with a context that holds no translator. A deployment's edits change what people see, never the logs.

### 5.3 Arguments

Each argument is prepared before formatting, in this order:

| Argument | For people (`T`) | For logs (`Error()`) |
| :---- | :---- | :---- |
| `nil`, or a nil pointer | `Unknown`'s text in an `error` parameter, else `""` | as `fmt` prints `nil` |
| an `error` | `T(ctx, arg)`: its public text, `Unknown`'s when it has none | `arg.Error()` |
| an `i18n.Ter` | `arg.T(ctx)` | `arg.T(context.Background())`, the code's text |
| anything else | unchanged | unchanged |

- Errors come before labels, because a public error is both.
- Error arguments are causes: `Unwrap` returns them (then `Wrap`'s cause), so `errors.Is` and `errors.As` see them.
- Numbers and dates go to `fmt` as they are. Locale-aware formatting is on i18n's roadmap.
- An occurrence's `T` ignores the `args` it's called with: `With` fixed its arguments. A `Def1`–`Def3`'s typed `T` prepares its arguments the same way, for a public error written by hand (§6.5).
- Render an error only through this package: `errors.T`, or an `error` parameter. Passed as an argument to a plain i18n key, it reaches `fmt`, which prints `Error()`.

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
- An occurrence has several causes (`Unwrap() []error`), so `errors.Unwrap` returns nil for it, as for `errors.Join`'s errors. Use `Is`, `As` or `AsType`.
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

`Key`–`Key3` also panic when given a kind twice, or a `Kind` that isn't one of the eight.

`Key`'s text is never formatted, so `%` is an ordinary character in it. A text that needs arguments is declared with `Key1`–`Key3`; more than three means the message should be split.

What the compiler doesn't catch: returning a `Def1`–`Def3` itself instead of an occurrence. It isn't public, so people see `Unknown`, and logs show its verbs. A test that checks the message catches it.

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

`All` yields nothing when no error in the tree is public; a page that must say something then shows `errors.T(ctx, err)`, which is `Unknown`'s text.

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

The status follows the error people see, so `errors.Join(sql.ErrConnDone, ErrOrderNotFound)` is a 404 and this helper doesn't log it. Code that joins a private error with public ones logs the private one where it joins them.

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

These are the contract's acceptance tests: 21 cases in 6 tests, plus the checker's 14 cases below. On 2026-11-14 all of them passed with `-race` against i18n `v0.2.0` through `i18ntest`, in a scratch spike built from [brief.md](brief.md) §8's sketches ([status.md](status.md)).

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

## 8. Out of v1

- Field paths on errors. Forms keep one error per field, or list them with `All`.
- Stack traces.
- An HTTP writer in the package. Apps own their response bodies; §6.3 shows one.
- The codes of composed inner errors in an API body, which carries the outer code only.
- A different text for `Unknown` inside a slot than alone.
- More than one message in a slot: an error argument shows `Public`'s.
- Plural forms and locale-aware numbers in arguments, which are on i18n's roadmap.
- A vet check for a `Def1`–`Def3` returned without its arguments.

## 9. The i18n amendment

Made in the i18n repository under its own terms; recorded here because D15 refers to it. This package doesn't wait for it.

i18n's interface §5.2 gains one rule, recorded in its §13:

> An argument that implements `Ter` is rendered with the same context before the text is formatted, so `Hello.T(ctx, settings.L10nServer)` shows the label in the request's locale. With no translator in the context, that's the label's code text.

- Today such an argument goes to `fmt` as it is, which prints a struct dump, because the key type has no `String` method.
- Render the arguments before `T` opens its read transaction, in `resolveCtx` ahead of `db.View`. bbolt warns that even a nested read-only transaction in one goroutine can deadlock: the inner one can wait for a writer that is waiting for the outer one.
- The fit rules don't change: the argument is a string by the time the text is formatted, so the code's text uses `%s` or `%v` for it.
- It's made in the i18n repository, under i18n's terms, and released as `v0.3.0`. This package prepares its own arguments, so it doesn't wait for it.
- **Done when:** tests show a label argument in the request's locale, and in the code's text without a translator; a label whose own text takes a label renders both in the locale; a stress test with concurrent `Set`s runs ten minutes clean under `-race`; the fit tests are unchanged.

## 10. Amendments

| Date | Change |
|------|--------|
| 2026-11-14 | Drafted from brief.md §§3–7; §8 from its §10, §9 from its §9 |
| 2026-11-14 | From [review 0001](reviews/0001-design.md), made in place before the lock: `Def1`–`Def3` take a typed `T` and aren't `i18n.Ter`s, so one returned bare shows `Unknown` (§4, §5.3, §5.8); nil and nil pointers (§5.1, §5.3); a nil `error` argument fills its slot with `Unknown`; a repeated or unknown `Kind` panics (§5.8); `errors.Unwrap` of an occurrence (§5.5); rendering errors only through this package (§5.3); `All` yielding nothing and `Join` with mixed kinds (§6.3); `KindOf(nil)` (§4) |
