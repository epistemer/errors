// Package errors gives errors two messages: Error() for logs, and T(ctx) for
// people, in their locale through github.com/epistemer/i18n.
//
// It is a superset of the standard library's errors package, so an app
// imports this one in its place.
//
// # Two messages
//
// Error() is the code's English text. It gains context with each wrap and
// may contain internals; it is for logs. T(ctx) is the text people see. It
// is picked from the error tree, not built up by wrapping: Public and T walk
// the tree in the order errors.As uses and take the first error that
// implements i18n.Ter, without looking inside it; All yields the outermost
// such error on each branch. Any wrapper works (%w, two
// %ws, Join, third-party wrappers) without knowing about i18n, and people
// never see Error() text.
//
// # Declaring
//
// Errors people see are declared in package-level variables. Each
// declaration registers an i18n key, so deployments can translate and edit
// its text:
//
//	var (
//		ErrOrderNotFound = errors.Key("errors.shop.order_not_found", "order not found", errors.NotFound)
//		ErrTooHigh       = errors.Key2[i18n.Ter, int]("errors.form.too_high", "%s must be at most %d", errors.Invalid)
//		ErrCheckoutFailed = errors.Key1[error]("errors.checkout.failed", "checkout failed: %s")
//	)
//
// A Def is its own sentinel. Key1, Key2 and Key3 declare texts with typed
// arguments, checked when they are declared; With makes an occurrence. An
// error argument composes: its public text fills the slot, and it is a
// cause. Wrap adds a cause people don't see.
//
// # Showing
//
// T(ctx, err) renders the error people see, All lists one per branch for
// forms, and KindOf gives its Kind and so its HTTP status. An error with no
// public part shows as Unknown.
//
// The contract is docs/interface.md in the repository.
package errors
