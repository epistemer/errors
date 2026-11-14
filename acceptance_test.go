// The acceptance tests of docs/interface.md §7, verbatim.

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
