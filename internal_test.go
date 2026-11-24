package errors

import (
	"reflect"
	"strings"
	"testing"

	"github.com/epistemer/i18n"
)

// The declaration checks (§5.8), on the checker itself: a declaration that
// panics can't be a package-level variable, and one that passes can't be
// made in a test after Load.
func TestCheckText(t *testing.T) {
	ter, i, e, a := reflect.TypeFor[i18n.Ter](), reflect.TypeFor[int](), reflect.TypeFor[error](), reflect.TypeFor[any]()
	tests := []struct {
		text string
		ts   []reflect.Type
		want string // "" when accepted, else a part of the panic
	}{
		{"%s is required", []reflect.Type{ter}, ""},
		{"100%% of %s", []reflect.Type{ter}, ""},
		{"%[2]d for %[1]s", []reflect.Type{ter, i}, ""},
		{"%v", []reflect.Type{a}, ""},
		{"%d", []reflect.Type{a}, ""},
		{"line %d: %s", []reflect.Type{i, e}, ""},
		{"no verbs", []reflect.Type{ter}, "doesn't use argument 1"},
		{"%s only", []reflect.Type{ter, i}, "doesn't use argument 2"},
		{"%s and %s", []reflect.Type{ter}, "can't format"},
		{"%s", []reflect.Type{i}, "doesn't suit"},
		{"%d", []reflect.Type{e}, "doesn't suit"},
		{"%*d", []reflect.Type{i}, "can't format"},
		{"%[3]s", []reflect.Type{ter}, "can't format"},
		{"50%", []reflect.Type{ter}, "can't format"},
	}
	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := panicOf(func() { checkText("errors.test.k", tt.text, tt.ts) })
			switch {
			case tt.want == "" && got != "":
				t.Errorf("panicked: %s", got)
			case tt.want != "" && !strings.Contains(got, tt.want):
				t.Errorf("panic = %q, want one containing %q", got, tt.want)
			case tt.want != "" && !strings.Contains(got, "errors.test.k"):
				t.Errorf("panic %q doesn't name the key", got)
			}
		})
	}
}

func TestDeclareRefusesBadKinds(t *testing.T) {
	for _, opts := range [][]Option{{Kind(0)}, {Kind(99)}, {NotFound, Conflict}, {NotFound, NotFound}} {
		if got := panicOf(func() { declare("errors.test.kinds", "t", opts, nil) }); !strings.Contains(got, "errors.test.kinds") {
			t.Errorf("declare(%v) panic = %q, want one naming the key", opts, got)
		}
	}
}

func TestMeta(t *testing.T) {
	ter, i := reflect.TypeFor[i18n.Ter](), reflect.TypeFor[int]()
	tests := []struct {
		k    Kind
		ts   []reflect.Type
		want string
	}{
		{0, nil, ""},
		{NotFound, nil, `{"kind":"not_found"}`},
		{0, []reflect.Type{reflect.TypeFor[error]()}, `{"args":["error"]}`},
		{Invalid, []reflect.Type{ter, i}, `{"kind":"invalid","args":["i18n.Ter","int"]}`},
	}
	for _, tt := range tests {
		for range 3 { // byte-stable
			if got := string(metaOf(tt.k, tt.ts)); got != tt.want {
				t.Errorf("metaOf(%v, %v) = %q, want %q", tt.k, tt.ts, got, tt.want)
			}
		}
	}
	if metaOf(0, nil) != nil {
		t.Error("a plain Key must have nil meta, to agree with a plain i18n.Key")
	}
}

func panicOf(f func()) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			msg, _ = r.(string)
			if msg == "" {
				msg = "non-string panic"
			}
		}
	}()
	f()
	return ""
}

func TestKindByName(t *testing.T) {
	tests := []struct {
		name string
		want Kind
	}{
		{"internal", Internal},
		{"invalid", Invalid},
		{"unauthenticated", Unauthenticated},
		{"forbidden", Forbidden},
		{"not_found", NotFound},
		{"conflict", Conflict},
		{"rate_limited", RateLimited},
		{"unavailable", Unavailable},
		{"unspecified", 0},
		{"", 0},
		{"NotFound", 0},
		{"bogus", 0},
	}
	for _, tt := range tests {
		if got := kindByName(tt.name); got != tt.want {
			t.Errorf("kindByName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}
