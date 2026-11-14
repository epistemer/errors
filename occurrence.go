package errors

import (
	"context"
)

// occurrence is an error made by With or Wrap: a declaration, its
// arguments, and a cause people don't see.
type occurrence struct {
	def   error // the declaration, for Is
	c     *core
	args  []any
	cause error
}

// Error returns the code's text formatted with the arguments as logs see
// them, then ": " and the cause's Error() when there is one. It renders with
// no translator, so a deployment's edits never reach logs.
func (o *occurrence) Error() string {
	s := o.c.h.T(context.Background(), forLogs(o.args)...)
	if !isNil(o.cause) {
		s += ": " + o.cause.Error()
	}
	return s
}

// K returns the declaration's key.
func (o *occurrence) K() string { return o.c.h.K() }

// T returns the text in ctx's locale, with the arguments With fixed. Its own
// arguments are ignored.
func (o *occurrence) T(ctx context.Context, _ ...any) string {
	return o.c.h.T(ctx, o.c.people(ctx, o.args)...)
}

// Kind returns the declaration's kind, 0 when it declared none.
func (o *occurrence) Kind() Kind { return o.c.kind }

// Is reports whether target is o's declaration.
func (o *occurrence) Is(target error) bool { return target == o.def }

// Unwrap returns the error arguments, then the cause.
func (o *occurrence) Unwrap() []error {
	out := o.shownErrors()
	if !isNil(o.cause) {
		out = append(out, o.cause)
	}
	return out
}

// shownErrors returns the error arguments, whose public text people see.
func (o *occurrence) shownErrors() []error {
	var out []error
	for _, a := range o.args {
		if e, ok := a.(error); ok && !isNil(e) {
			out = append(out, e)
		}
	}
	return out
}
