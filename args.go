package errors

import (
	"context"
	"reflect"

	"github.com/epistemer/i18n"
)

// people prepares arguments for people: an error shows as its public text, a
// label in the same locale. An error's Error() never reaches them. A nil, or
// a nil pointer, fills an error parameter's slot with Unknown's text, and
// any other slot with "".
func (c *core) people(ctx context.Context, args []any) []any {
	if len(args) == 0 {
		return nil
	}
	out := make([]any, len(args))
	for i, a := range args {
		if isNil(a) {
			if c.ts[i].Implements(errorType) {
				out[i] = Unknown.T(ctx)
			} else {
				out[i] = ""
			}
			continue
		}
		switch v := a.(type) {
		case error:
			out[i] = T(ctx, v)
		case i18n.Ter:
			out[i] = v.T(ctx)
		default:
			out[i] = a
		}
	}
	return out
}

// forLogs prepares arguments for Error(): an error as its Error(), a label
// as the code's text, and a nil pointer as nil.
func forLogs(args []any) []any {
	if len(args) == 0 {
		return nil
	}
	out := make([]any, len(args))
	for i, a := range args {
		if isNil(a) {
			out[i] = nil
			continue
		}
		switch v := a.(type) {
		case error:
			out[i] = v.Error()
		case i18n.Ter:
			out[i] = v.T(context.Background())
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
