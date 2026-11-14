package errors

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/epistemer/i18n"
)

var (
	terType   = reflect.TypeFor[i18n.Ter]()
	errorType = reflect.TypeFor[error]()
	badFmt    = regexp.MustCompile(`%!.?\(`)
)

// probe marks where fmt puts an argument.
type probe int

func (p probe) Format(f fmt.State, _ rune) { fmt.Fprintf(f, "\x00%d\x00", int(p)) }

// checkText panics unless text uses each argument, and fmt can format it
// with values of the argument types as T will see them. fmt is the parser.
func checkText(key, text string, ts []reflect.Type) {
	probes := make([]any, len(ts))
	for i := range ts {
		probes[i] = probe(i)
	}
	out := fmt.Sprintf(text, probes...)
	// fmt appends unused arguments as %!(EXTRA …); report them as unused.
	if i := strings.Index(out, "%!(EXTRA "); i >= 0 {
		out = out[:i]
	}
	if badFmt.MatchString(out) {
		panic(fmt.Sprintf("errors: %s: fmt can't format %q with %d arguments", key, text, len(ts)))
	}
	for i := range ts {
		if !strings.Contains(out, fmt.Sprintf("\x00%d\x00", i)) {
			panic(fmt.Sprintf("errors: %s: %q doesn't use argument %d", key, text, i+1))
		}
	}
	zeros := make([]any, len(ts))
	for i, t := range ts {
		switch {
		case t.Implements(errorType), t.Implements(terType):
			zeros[i] = "" // T turns these into text
		case t.Kind() == reflect.Interface:
			zeros[i] = probe(i) // the dynamic type isn't known
		default:
			zeros[i] = reflect.Zero(t).Interface()
		}
	}
	if out := fmt.Sprintf(text, zeros...); badFmt.MatchString(out) {
		panic(fmt.Sprintf("errors: %s: a verb in %q doesn't suit its argument's type: %s", key, text, out))
	}
}

// meta is a declaration's i18n meta: its kind and argument types.
type meta struct {
	Kind string   `json:"kind,omitempty"`
	Args []string `json:"args,omitempty"`
}

// metaOf returns a declaration's i18n meta, nil without a kind and
// arguments, so a plain Key agrees with a plain i18n.Key.
func metaOf(k Kind, ts []reflect.Type) []byte {
	var m meta
	if k != 0 {
		m.Kind = k.String()
	}
	for _, t := range ts {
		m.Args = append(m.Args, t.String())
	}
	if m.Kind == "" && len(m.Args) == 0 {
		return nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		panic(err) // strings only: can't happen
	}
	return b
}
