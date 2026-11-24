package errors

import (
	"context"
	"iter"

	"github.com/epistemer/i18n"
)

// Unknown is what people see for an error with no public part. It is a key
// like any other, so deployments can edit its text.
var Unknown = Key("errors.unknown", "something went wrong", Internal)

// Public returns the first error in err's tree, in the order errors.As
// uses, that implements i18n.Ter; Unknown when there is none; nil for nil.
// The walk never looks inside a public error. One without a key has the key
// "".
func Public(err error) i18n.TKer {
	if isNil(err) {
		return nil
	}
	if n := first(err); n != nil {
		return asTKer(n)
	}
	return Unknown
}

// T returns the text people see for err in ctx's locale: Public(err).T(ctx),
// and "" for nil.
func T(ctx context.Context, err error) string {
	if isNil(err) {
		return ""
	}
	return Public(err).T(ctx)
}

// All yields the outermost public error on each branch of err's tree, in
// the order errors.As uses. When it yields anything, its first value is
// Public(err). It yields nothing when no error in the tree is public.
func All(err error) iter.Seq[i18n.TKer] {
	return func(yield func(i18n.TKer) bool) {
		walk(err, func(n error) bool { return yield(asTKer(n)) })
	}
}

// KindOf returns the kind of Public(err): its own when it declares one
// (from a Kind() Kind method returning non-zero, or a Kind() string method
// returning one of the eight kind names); else, for an occurrence with error
// arguments, KindOf its first; else Internal. KindOf(nil) is Internal. A cause
// given to Wrap never lends its kind.
func KindOf(err error) Kind {
	n := first(err)
	if n == nil {
		return Internal
	}
	if k, ok := n.(interface{ Kind() Kind }); ok && k.Kind() != 0 {
		return k.Kind()
	}
	if k, ok := n.(interface{ Kind() string }); ok {
		if kind := kindByName(k.Kind()); kind != 0 {
			return kind
		}
	}
	if o, ok := n.(interface{ shownErrors() []error }); ok {
		if errs := o.shownErrors(); len(errs) > 0 {
			return KindOf(errs[0])
		}
	}
	return Internal
}

// first returns the first public error in err's tree, or nil.
func first(err error) (found error) {
	walk(err, func(n error) bool { found = n; return false })
	return found
}

// walk calls f on each public error that no other public error wraps, in
// pre-order, depth-first; it doesn't descend into a public error. It stops
// when f returns false, and returns false then.
func walk(err error, f func(error) bool) bool {
	if isNil(err) {
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

// noKey gives a public error without a key the key "".
type noKey struct{ i18n.Ter }

func (noKey) K() string { return "" }

func asTKer(n error) i18n.TKer {
	if tk, ok := n.(i18n.TKer); ok {
		return tk
	}
	return noKey{n.(i18n.Ter)}
}
