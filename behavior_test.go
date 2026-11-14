package errors_test

import (
	"context"
	"database/sql"
	stderrors "errors"
	"fmt"
	"io/fs"
	"testing"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
)

// M1.1: the standard library's API behaves as the standard library's.
func TestStandardAPI(t *testing.T) {
	if errors.ErrUnsupported != stderrors.ErrUnsupported {
		t.Error("ErrUnsupported isn't the standard library's")
	}
	e := errors.New("boom")
	if e.Error() != "boom" || e == errors.New("boom") {
		t.Error("New must return a distinct error with its text")
	}
	w := fmt.Errorf("ctx: %w", fs.ErrNotExist)
	if !errors.Is(w, fs.ErrNotExist) || errors.Is(w, fs.ErrExist) {
		t.Error("Is")
	}
	if errors.Unwrap(w) != fs.ErrNotExist || errors.Unwrap(e) != nil {
		t.Error("Unwrap")
	}
	pe := &fs.PathError{Op: "open", Path: "x", Err: fs.ErrNotExist}
	var target *fs.PathError
	if !errors.As(fmt.Errorf("a: %w", pe), &target) || target != pe {
		t.Error("As")
	}
	if got, ok := errors.AsType[*fs.PathError](fmt.Errorf("a: %w", pe)); !ok || got != pe {
		t.Error("AsType")
	}
	if _, ok := errors.AsType[*fs.PathError](e); ok {
		t.Error("AsType found what isn't there")
	}
	j := errors.Join(e, nil, fs.ErrNotExist)
	if j.Error() != stderrors.Join(e, fs.ErrNotExist).Error() || !errors.Is(j, fs.ErrNotExist) {
		t.Error("Join")
	}
	if errors.Join() != nil || errors.Join(nil) != nil {
		t.Error("Join of nothing must be nil")
	}
}

// M1.2: every kind's name and status, and the zero Kind.
func TestKinds(t *testing.T) {
	tests := []struct {
		k      errors.Kind
		name   string
		status int
	}{
		{0, "unspecified", 500},
		{errors.Internal, "internal", 500},
		{errors.Invalid, "invalid", 400},
		{errors.Unauthenticated, "unauthenticated", 401},
		{errors.Forbidden, "forbidden", 403},
		{errors.NotFound, "not_found", 404},
		{errors.Conflict, "conflict", 409},
		{errors.RateLimited, "rate_limited", 429},
		{errors.Unavailable, "unavailable", 503},
	}
	for _, tt := range tests {
		if tt.k.String() != tt.name || tt.k.Status() != tt.status {
			t.Errorf("Kind %d = %q %d, want %q %d", tt.k, tt.k.String(), tt.k.Status(), tt.name, tt.status)
		}
	}
	var _ errors.Option = errors.NotFound
}

// M1.4: Error() stays the code's text with an Arabic edit and an English
// edit in the store.
func TestErrorIgnoresEdits(t *testing.T) {
	ar, en := contexts(t)
	tests := []struct {
		err  error
		want string
	}{
		{ErrOrderNotFound, "order not found"},
		{ErrOrderNotSaved.Wrap(ErrOrderNotFound), "the order couldn't be saved: order not found"},
		{ErrRequired.With(L10nName), "Name is required"},
		{ErrCheckoutFailed.With(ErrPaymentDeclined.With(ErrLowBalance)), "checkout failed: payment declined: insufficient balance"},
	}
	for _, tt := range tests {
		if got := tt.err.Error(); got != tt.want {
			t.Errorf("Error() = %q, want %q", got, tt.want)
		}
	}
	// The edits are there: people see them.
	if errors.T(en, ErrOrderNotFound) != "We couldn't find that order" || errors.T(ar, ErrOrderNotFound) != "الطلب غير موجود" {
		t.Error("the edits aren't in the store")
	}
}

var (
	ErrThree = errors.Key3[int, i18n.Ter, error]("errors.test.three", "row %d, %s: %s")
	ErrDup   = errors.Key("errors.test.dup", "duplicate", errors.Conflict)
)

type ptrErr struct{}

func (*ptrErr) Error() string { return "ptr" }

type ptrTer struct{}

func (*ptrTer) Error() string                    { return "ptr" }
func (*ptrTer) T(context.Context, ...any) string { return "never" }

// Review 0001's changes to the interface (§10).
func TestReviewChanges(t *testing.T) {
	ar, _ := contexts(t)
	const unknown = "حدث خطأ ما"

	t.Run("a bare Def1-Def3 isn't public", func(t *testing.T) {
		if _, ok := any(ErrTooHigh).(i18n.Ter); ok {
			t.Fatal("Def2 is an i18n.Ter")
		}
		for _, err := range []error{ErrTooHigh, ErrRequired, ErrThree, fmt.Errorf("x: %w", ErrCheckoutFailed)} {
			if got := errors.T(ar, err); got != unknown {
				t.Errorf("T(%v) = %q, want Unknown", err, got)
			}
			if got := errors.KindOf(err); got != errors.Internal {
				t.Errorf("KindOf(%v) = %v", err, got)
			}
		}
		if got := ErrTooHigh.Error(); got != "%s must be at most %d" {
			t.Errorf("logs see %q, want the verbs", got)
		}
	})

	t.Run("typed T", func(t *testing.T) {
		if got := ErrTooHigh.T(ar, L10nAmount, 3); got != "يجب ألا يتجاوز المبلغ 3" {
			t.Errorf("Def2.T = %q", got)
		}
		if got := ErrPaymentDeclined.T(ar, ErrLowBalance); got != "رُفض الدفع: الرصيد غير كافٍ" {
			t.Errorf("Def1.T = %q", got)
		}
	})

	t.Run("Key3", func(t *testing.T) {
		err := ErrThree.Wrap(sql.ErrConnDone, 4, L10nAmount, ErrDup)
		if got, want := err.Error(), "row 4, Amount: duplicate: sql: connection is already closed"; got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
		if got, want := errors.T(ar, err), "row 4, المبلغ: duplicate"; got != want {
			t.Errorf("T = %q, want %q", got, want)
		}
		if errors.KindOf(err) != errors.Conflict || !errors.Is(err, ErrThree) || !errors.Is(err, ErrDup) || !errors.Is(err, sql.ErrConnDone) {
			t.Error("kind or identity")
		}
	})

	t.Run("nil and nil pointers", func(t *testing.T) {
		var p *ptrErr
		var pt *ptrTer
		for _, e := range []error{nil, p, pt} {
			err := ErrPaymentDeclined.With(e)
			if got := errors.T(ar, err); got != "رُفض الدفع: "+unknown {
				t.Errorf("With(%#v): T = %q", e, got)
			}
			_ = err.Error()
			if errors.KindOf(err) != errors.Internal || errors.Is(err, ErrLowBalance) {
				t.Errorf("With(%#v): kind or identity", e)
			}
			if got := ErrOrderNotSaved.Wrap(e); got != error(ErrOrderNotSaved) {
				t.Errorf("Wrap(%#v) = %v, want the declaration", e, got)
			}
			if errors.Public(e) != nil || errors.T(ar, e) != "" {
				t.Errorf("Public(%#v) must be nil", e)
			}
		}
		if got := ErrRequired.With(nil); errors.T(ar, got) != " مطلوب" {
			t.Errorf("a nil label: T = %q", errors.T(ar, got))
		}
	})

	t.Run("Unwrap of an occurrence", func(t *testing.T) {
		if errors.Unwrap(ErrOrderNotSaved.Wrap(sql.ErrConnDone)) != nil {
			t.Error("an occurrence has several causes; Unwrap must be nil")
		}
	})
}
