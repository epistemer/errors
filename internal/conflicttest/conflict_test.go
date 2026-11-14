// Package conflicttest shows i18n's Load refusing one key declared with two
// kinds (interface §5.9). It is a test binary of its own, because the
// conflict makes Load fail for every key the binary registers.
package conflicttest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/epistemer/errors"
	"github.com/epistemer/i18n"
	"github.com/epistemer/i18n/i18ntest"
)

var (
	_ = errors.Key("errors.test.twice", "twice", errors.NotFound)
	_ = errors.Key("errors.test.twice", "twice", errors.Conflict)

	_ = errors.Key1[int]("errors.test.args", "%v")
	_ = errors.Key1[string]("errors.test.args", "%v")
)

// fatal records the failure i18ntest reports instead of failing the test.
type fatal struct {
	*testing.T
	msg string
}

func (f *fatal) Fatalf(format string, args ...any) { f.msg = fmt.Sprintf(format, args...) }

func TestLoadRefusesTwoMeanings(t *testing.T) {
	f := &fatal{T: t}
	i18ntest.New(f, i18n.Locale{Code: "en", NativeName: "English"})
	for _, key := range []string{"errors.test.twice", "errors.test.args"} {
		want := fmt.Sprintf("key %q registered twice, with different meta", key)
		if !strings.Contains(f.msg, want) {
			t.Errorf("Load's error %q doesn't contain %q", f.msg, want)
		}
	}
}
