# errors

neozp's error package. An error carries two messages, and neither is derived
from the other:

- `Error()` is for logs: the code's English text, gaining context with each
  wrap.
- `T(ctx)` is for people: text in their locale, through
  [i18n](https://github.com/epistemer/i18n), picked from the error tree
  rather than built up by wrapping, so `%w`, `errors.Join` and third-party
  wrappers all keep working, and people never see `Error()` text.

It is a superset of the standard library's `errors`, so it replaces that
import. Declarations are typed and checked when declared, an error argument
composes into the outer message, and every error people see has a kind and
so an HTTP status.

```sh
go get github.com/epistemer/errors
```

## Quickstart

Declare the errors people see once, at package level. Each is an i18n key,
so a deployment translates and edits it in the running app:

```go
var (
	ErrOrderNotFound  = errors.Key("errors.shop.order_not_found", "order not found", errors.NotFound)
	ErrLowBalance     = errors.Key("errors.wallet.low_balance", "insufficient balance", errors.Conflict)
	ErrCheckoutFailed = errors.Key1[error]("errors.checkout.failed", "checkout failed: %s")
)
```

Return them, wrap them, or compose them, and show them where the request
ends. With i18n set up and a few Arabic translations in its store (the full
program is [below](#the-quickstart-program)):

```go
for _, err := range []error{
	fmt.Errorf("load order 42: %w", ErrOrderNotFound), // context for logs only
	ErrCheckoutFailed.With(ErrLowBalance),               // composed: one message inside another
	fmt.Errorf("query orders: %w", sql.ErrConnDone),     // nothing public: people see Unknown
} {
	fmt.Printf("log:    %v\npeople: %s (%s, %d)\n", err, errors.T(ctx, err),
		errors.Public(err).K(), errors.KindOf(err).Status())
}
```

Output, from a real run:

```
log:    load order 42: order not found
people: الطلب غير موجود (errors.shop.order_not_found, 404)
log:    checkout failed: insufficient balance
people: تعذر إتمام الطلب: الرصيد غير كافٍ (errors.checkout.failed, 409)
log:    query orders: sql: connection is already closed
people: حدث خطأ ما (errors.unknown, 500)
```

- `errors.T(ctx, err)` is the text people see; `errors.All(err)` lists one
  per branch, for forms.
- `errors.Public(err).K()` is a stable code for an API, and
  `errors.KindOf(err).Status()` its HTTP status.
- `Wrap` keeps a cause people don't see: `ErrOrderNotSaved.Wrap(err)` logs
  the driver's error and keeps it for `errors.Is`, and shows people only its
  own text.
- Texts are lowercase with no final period, like Go error strings, so they
  compose. Capitalize a message shown alone in the UI, with CSS
  `::first-letter`.

The [package docs](https://pkg.go.dev/github.com/epistemer/errors) have
runnable examples, and [docs/interface.md](docs/interface.md) is the full
contract.

### The quickstart program

```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	bolt "go.etcd.io/bbolt"
)

var (
	ErrOrderNotFound  = errors.Key("errors.shop.order_not_found", "order not found", errors.NotFound)
	ErrLowBalance     = errors.Key("errors.wallet.low_balance", "insufficient balance", errors.Conflict)
	ErrCheckoutFailed = errors.Key1[error]("errors.checkout.failed", "checkout failed: %s")
)

func main() {
	db, err := bolt.Open(filepath.Join(os.TempDir(), "quickstart.db"), 0o600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	tr := i18n.New(db, i18n.Locale{Code: "en", NativeName: "English"}, nil)
	if err := tr.Load(); err != nil {
		log.Fatal(err)
	}
	// Translations are normally typed into the running app; seeded here.
	if err := tr.CreateLocale(i18n.Locale{Code: "ar", NativeName: "العربية", RTL: true}); err == nil {
		tr.Set("ar", "errors.shop.order_not_found", "الطلب غير موجود")
		tr.Set("ar", "errors.wallet.low_balance", "الرصيد غير كافٍ")
		tr.Set("ar", "errors.checkout.failed", "تعذر إتمام الطلب: %s")
		tr.Set("ar", "errors.unknown", "حدث خطأ ما")
	}
	ar, _ := tr.Locale("ar")
	ctx := i18n.WithContext(context.Background(), tr, ar)

	for _, err := range []error{
		fmt.Errorf("load order 42: %w", ErrOrderNotFound),
		ErrCheckoutFailed.With(ErrLowBalance),
		fmt.Errorf("query orders: %w", sql.ErrConnDone),
	} {
		fmt.Printf("log:    %v\npeople: %s (%s, %d)\n", err, errors.T(ctx, err),
			errors.Public(err).K(), errors.KindOf(err).Status())
	}
}
```

## Documents

- [docs/interface.md](docs/interface.md): the contract, locked 2026-11-14.
- [ADR 0001](docs/adr/0001-two-messages.md): the design and why.
- [docs/roadmap.md](docs/roadmap.md) and [docs/status.md](docs/status.md):
  the plan and the log of working sessions.
- [docs/brief.md](docs/brief.md): the design brief the work started from.

Part of neozp, alongside [cfgx](https://github.com/epistemer/cfgx),
[i18n](https://github.com/epistemer/i18n),
[genx](https://github.com/epistemer/genx) and
[models](https://github.com/epistemer/models). License: [MIT](LICENSE).
