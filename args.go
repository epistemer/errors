package errors

import (
	"context"
	"reflect"

	"github.com/epistemer/i18n"
)

// people prepares arguments for people: an error shows as its public text, a
// label in the same locale. An error's Error() never reaches them. A nil in
// an error parameter fills its slot with Unknown's text; other nils are "".
func (c *core) people(ctx context.Context, args []any) []any {
	if len(args) == 0 {
		return nil
	}
	out := make([]any, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case error:
			if isNil(v) {
				out[i] = Unknown.T(ctx)
			} else {
				out[i] = T(ctx, v)
			}
		case i18n.Ter:
			if isNil(v) {
				out[i] = ""
			} else {
				out[i] = v.T(ctx)
			}
		case nil:
			if c.ts[i].Implements(errorType) {
				out[i] = Unknown.T(ctx)
			} else {
				out[i] = ""
			}
		default:
			out[i] = a
		}
	}
	return out
}

// forLogs prepares arguments for Error(): an error as its Error(), a label
// as the code's text.
func forLogs(args []any) []any {
	if len(args) == 0 {
		return nil
	}
	out := make([]any, len(args))
	for i, a := range args {
		switch v := a.(type) {
		case error:
			if isNil(v) {
				out[i] = nil
			} else {
				out[i] = v.Error()
			}
		case i18n.Ter:
			if isNil(v) {
				out[i] = nil
			} else {
				out[i] = v.T(context.Background())
			}
		default:
			out[i] = a
		}
	}
	return out
}

// isNil reports whether v is nil, or a nil pointer in an interface.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	r := reflect.ValueOf(v)
	return r.Kind() == reflect.Pointer && r.IsNil()
}
