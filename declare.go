package errors

import (
	"context"
	"fmt"
	"reflect"

	"github.com/epistemer/i18n"
)

// core is what every declaration holds: its i18n handle, the code's text,
// its kind, and its argument types.
type core struct {
	h    i18n.TKer
	text string
	kind Kind
	ts   []reflect.Type
}

// declare checks a declaration and registers its key with i18n.
func declare(key, text string, opts []Option, ts []reflect.Type) core {
	var k Kind
	for _, o := range opts {
		kk, ok := o.(Kind)
		if !ok {
			continue
		}
		switch {
		case !kk.valid():
			panic(fmt.Sprintf("errors: %s: %v isn't a kind", key, kk))
		case k != 0:
			panic(fmt.Sprintf("errors: %s: two kinds, %v and %v", key, k, kk))
		}
		k = kk
	}
	if len(ts) > 0 {
		checkText(key, text, ts)
	}
	return core{h: i18n.Key(key, text, metaOf(k, ts)...), text: text, kind: k, ts: ts}
}

// Def declares an error people see, with a text that takes no arguments. It
// is its own sentinel: return it, and test for it with Is.
type Def struct{ c core }

// Key declares an error whose text takes no arguments: key is its i18n key,
// text the code's English text, never formatted. Call it only in
// package-level variable initializers: it calls i18n.Key, which panics after
// i18n's Load.
func Key(key, text string, opts ...Option) *Def {
	return &Def{declare(key, text, opts, nil)}
}

// Error returns the code's text.
func (d *Def) Error() string { return d.c.text }

// K returns the key, which is also the error's code in an API.
func (d *Def) K() string { return d.c.h.K() }

// T returns the text in ctx's locale. Its arguments are ignored.
func (d *Def) T(ctx context.Context, _ ...any) string { return d.c.h.T(ctx) }

// Kind returns the kind as declared, and 0 when none was.
func (d *Def) Kind() Kind { return d.c.kind }

// Wrap returns d over a cause people don't see: people see d's text, logs
// see d's text, ": " and the cause's. Wrap(nil) is d.
func (d *Def) Wrap(cause error) error {
	if isNil(cause) {
		return d
	}
	return &occurrence{def: d, c: &d.c, cause: cause}
}

// Def1 declares an error people see, with a text that takes one argument.
// It is a sentinel for its occurrences, which With and Wrap make. It isn't
// an i18n.Ter: returned itself, people see Unknown.
type Def1[A any] struct{ c core }

// Key1 declares an error whose text takes one argument, of the types given.
// It panics unless fmt can format text with them and it uses each one.
// Call it only in package-level variable initializers: it calls i18n.Key,
// which panics after i18n's Load.
func Key1[A any](key, text string, opts ...Option) *Def1[A] {
	return &Def1[A]{declare(key, text, opts, []reflect.Type{reflect.TypeFor[A]()})}
}

// Error returns the code's text, unformatted.
func (d *Def1[A]) Error() string { return d.c.text }

// K returns the key, which is also the error's code in an API.
func (d *Def1[A]) K() string { return d.c.h.K() }

// T returns the text in ctx's locale, formatted with the arguments: an
// error shows as its public text, an i18n.Ter in ctx's locale.
func (d *Def1[A]) T(ctx context.Context, a A) string {
	return d.c.h.T(ctx, d.c.people(ctx, []any{a})...)
}

// Kind returns the kind as declared, and 0 when none was.
func (d *Def1[A]) Kind() Kind { return d.c.kind }

// With returns an occurrence of d with these arguments. An error argument
// is a cause.
func (d *Def1[A]) With(a A) error {
	return &occurrence{def: d, c: &d.c, args: []any{a}}
}

// Wrap returns an occurrence of d with these arguments, over a cause people
// don't see.
func (d *Def1[A]) Wrap(cause error, a A) error {
	return &occurrence{def: d, c: &d.c, args: []any{a}, cause: cause}
}

// Def2 declares an error people see, with a text that takes two arguments.
// It is a sentinel for its occurrences, which With and Wrap make. It isn't
// an i18n.Ter: returned itself, people see Unknown.
type Def2[A, B any] struct{ c core }

// Key2 declares an error whose text takes two arguments, of the types given.
// It panics unless fmt can format text with them and it uses each one.
// Call it only in package-level variable initializers: it calls i18n.Key,
// which panics after i18n's Load.
func Key2[A, B any](key, text string, opts ...Option) *Def2[A, B] {
	return &Def2[A, B]{declare(key, text, opts, []reflect.Type{reflect.TypeFor[A](), reflect.TypeFor[B]()})}
}

// Error returns the code's text, unformatted.
func (d *Def2[A, B]) Error() string { return d.c.text }

// K returns the key, which is also the error's code in an API.
func (d *Def2[A, B]) K() string { return d.c.h.K() }

// T returns the text in ctx's locale, formatted with the arguments: an
// error shows as its public text, an i18n.Ter in ctx's locale.
func (d *Def2[A, B]) T(ctx context.Context, a A, b B) string {
	return d.c.h.T(ctx, d.c.people(ctx, []any{a, b})...)
}

// Kind returns the kind as declared, and 0 when none was.
func (d *Def2[A, B]) Kind() Kind { return d.c.kind }

// With returns an occurrence of d with these arguments. An error argument
// is a cause.
func (d *Def2[A, B]) With(a A, b B) error {
	return &occurrence{def: d, c: &d.c, args: []any{a, b}}
}

// Wrap returns an occurrence of d with these arguments, over a cause people
// don't see.
func (d *Def2[A, B]) Wrap(cause error, a A, b B) error {
	return &occurrence{def: d, c: &d.c, args: []any{a, b}, cause: cause}
}

// Def3 declares an error people see, with a text that takes three arguments.
// It is a sentinel for its occurrences, which With and Wrap make. It isn't
// an i18n.Ter: returned itself, people see Unknown.
type Def3[A, B, C any] struct{ c core }

// Key3 declares an error whose text takes three arguments, of the types given.
// It panics unless fmt can format text with them and it uses each one.
// Call it only in package-level variable initializers: it calls i18n.Key,
// which panics after i18n's Load.
func Key3[A, B, C any](key, text string, opts ...Option) *Def3[A, B, C] {
	return &Def3[A, B, C]{declare(key, text, opts, []reflect.Type{reflect.TypeFor[A](), reflect.TypeFor[B](), reflect.TypeFor[C]()})}
}

// Error returns the code's text, unformatted.
func (d *Def3[A, B, C]) Error() string { return d.c.text }

// K returns the key, which is also the error's code in an API.
func (d *Def3[A, B, C]) K() string { return d.c.h.K() }

// T returns the text in ctx's locale, formatted with the arguments: an
// error shows as its public text, an i18n.Ter in ctx's locale.
func (d *Def3[A, B, C]) T(ctx context.Context, a A, b B, c C) string {
	return d.c.h.T(ctx, d.c.people(ctx, []any{a, b, c})...)
}

// Kind returns the kind as declared, and 0 when none was.
func (d *Def3[A, B, C]) Kind() Kind { return d.c.kind }

// With returns an occurrence of d with these arguments. An error argument
// is a cause.
func (d *Def3[A, B, C]) With(a A, b B, c C) error {
	return &occurrence{def: d, c: &d.c, args: []any{a, b, c}}
}

// Wrap returns an occurrence of d with these arguments, over a cause people
// don't see.
func (d *Def3[A, B, C]) Wrap(cause error, a A, b B, c C) error {
	return &occurrence{def: d, c: &d.c, args: []any{a, b, c}, cause: cause}
}
