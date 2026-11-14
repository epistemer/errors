package errors_test

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
)

// A toy shop's errors, declared once at package level. Each is an i18n key,
// so a deployment can translate and edit its text.
var (
	ErrNoSuchOrder = errors.Key("errors.example.no_such_order", "order not found", errors.NotFound)
	ErrNotSaved    = errors.Key("errors.example.not_saved", "the order couldn't be saved")
	ErrNoFunds     = errors.Key("errors.example.no_funds", "insufficient balance", errors.Conflict)

	ErrOverLimit = errors.Key2[i18n.Ter, int]("errors.example.over_limit", "%s must be at most %d", errors.Invalid)
	ErrDeclined  = errors.Key1[error]("errors.example.declined", "payment declined: %s")

	L10nQuantity = i18n.Key("example.quantity", "Quantity")
)

// show prints what logs see, what people see, the code and the status.
func show(ctx context.Context, err error) {
	p, kind := errors.Public(err), errors.KindOf(err)
	fmt.Printf("log:    %s\npeople: %s\ncode:   %s, %d\n\n", err, p.T(ctx), p.K(), kind.Status())
}

// Wrapping adds context for logs; people see the declared message, in their
// locale when the context carries an i18n translator.
func Example() {
	ctx := context.Background() // no translator: the code's English text

	show(ctx, fmt.Errorf("load order 42: %w", ErrNoSuchOrder))
	show(ctx, ErrNotSaved.Wrap(sql.ErrConnDone))
	show(ctx, fmt.Errorf("query orders: %w", sql.ErrConnDone))
	show(ctx, ErrDeclined.With(ErrNoFunds))
	show(ctx, ErrOverLimit.With(L10nQuantity, 10))

	// Output:
	// log:    load order 42: order not found
	// people: order not found
	// code:   errors.example.no_such_order, 404
	//
	// log:    the order couldn't be saved: sql: connection is already closed
	// people: the order couldn't be saved
	// code:   errors.example.not_saved, 500
	//
	// log:    query orders: sql: connection is already closed
	// people: something went wrong
	// code:   errors.unknown, 500
	//
	// log:    payment declined: insufficient balance
	// people: payment declined: insufficient balance
	// code:   errors.example.declined, 409
	//
	// log:    Quantity must be at most 10
	// people: Quantity must be at most 10
	// code:   errors.example.over_limit, 400
}

// An error argument composes: its public text fills the slot, translators
// place it, and it stays a cause for Is and KindOf.
func ExampleDef1_With() {
	err := ErrDeclined.With(fmt.Errorf("debit wallet 7: %w", ErrNoFunds))

	fmt.Println(err)
	fmt.Println(errors.T(context.Background(), err))
	fmt.Println(errors.Is(err, ErrNoFunds), errors.KindOf(err))
	// Output:
	// payment declined: debit wallet 7: insufficient balance
	// payment declined: insufficient balance
	// true conflict
}

// Wrap keeps a cause for logs and errors.Is, and people don't see it.
func ExampleDef_Wrap() {
	err := ErrNotSaved.Wrap(context.DeadlineExceeded)

	fmt.Println(err)
	fmt.Println(errors.T(context.Background(), err))
	fmt.Println(errors.Is(err, ErrNotSaved), errors.Is(err, context.DeadlineExceeded))
	// Output:
	// the order couldn't be saved: context deadline exceeded
	// the order couldn't be saved
	// true true
}

// All lists one message per branch, for a form that shows every problem.
func ExampleAll() {
	err := fmt.Errorf("validate: %w", errors.Join(
		ErrOverLimit.With(L10nQuantity, 10),
		errors.New("audit log unavailable"), // private: not listed
		ErrNoSuchOrder,
	))
	for p := range errors.All(err) {
		fmt.Println(p.T(context.Background()))
	}
	// Output:
	// Quantity must be at most 10
	// order not found
}

// The status follows the error people see; a frame without a kind takes its
// first error argument's.
func ExampleKindOf() {
	for _, err := range []error{
		ErrNoSuchOrder,
		ErrDeclined.With(ErrNoFunds),
		ErrNotSaved.Wrap(ErrNoSuchOrder), // a Wrap cause lends nothing
		sql.ErrConnDone,
	} {
		fmt.Println(errors.KindOf(err), errors.KindOf(err).Status())
	}
	// Output:
	// not_found 404
	// conflict 409
	// internal 500
	// internal 500
}
