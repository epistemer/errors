package errors

import stderrors "errors"

// The standard library's API, so an app imports one errors package.

// ErrUnsupported is the standard library's [errors.ErrUnsupported].
var ErrUnsupported = stderrors.ErrUnsupported

// New returns a private error whose Error() is text. People see Unknown in
// its place.
func New(text string) error { return stderrors.New(text) }

// Is reports whether any error in err's tree matches target, as the standard
// library's [errors.Is] does. An occurrence matches its declaration.
func Is(err, target error) bool { return stderrors.Is(err, target) }

// As finds the first error in err's tree that matches target, as the
// standard library's [errors.As] does.
func As(err error, target any) bool { return stderrors.As(err, target) }

// AsType finds the first error in err's tree of type E, as the standard
// library's [errors.AsType] does.
func AsType[E error](err error) (E, bool) { return stderrors.AsType[E](err) }

// Unwrap returns the result of err's Unwrap method, as the standard
// library's [errors.Unwrap] does. It is nil for an error whose Unwrap
// returns a slice, as Join's errors and every occurrence do, even one with a
// single cause: use Is, As or AsType.
func Unwrap(err error) error { return stderrors.Unwrap(err) }

// Join returns an error that wraps errs, as the standard library's
// [errors.Join] does.
func Join(errs ...error) error { return stderrors.Join(errs...) }
