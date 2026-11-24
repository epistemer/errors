package errors

import "strconv"

// Kind classifies an error people see, and gives its HTTP status. The zero
// Kind means none was declared.
type Kind uint8

// The kinds. Each is an Option of a declaration.
const (
	Internal        Kind = iota + 1 // 500: the server's fault; Unknown
	Invalid                         // 400: the input is wrong
	Unauthenticated                 // 401: the caller isn't signed in
	Forbidden                       // 403: the caller may not do this
	NotFound                        // 404: no such thing
	Conflict                        // 409: the current state doesn't allow it
	RateLimited                     // 429: too many, too soon
	Unavailable                     // 503: try again later
)

var kinds = [...]struct {
	name   string
	status int
}{
	{"unspecified", 500},
	Internal:        {"internal", 500},
	Invalid:         {"invalid", 400},
	Unauthenticated: {"unauthenticated", 401},
	Forbidden:       {"forbidden", 403},
	NotFound:        {"not_found", 404},
	Conflict:        {"conflict", 409},
	RateLimited:     {"rate_limited", 429},
	Unavailable:     {"unavailable", 503},
}

func (k Kind) valid() bool { return k != 0 && int(k) < len(kinds) }

// kindByName returns the Kind for name as Kind.String() spells it, or 0
// when none matches (including "unspecified").
func kindByName(name string) Kind {
	for k := Kind(1); int(k) < len(kinds); k++ {
		if kinds[k].name == name {
			return k
		}
	}
	return 0
}

// Status returns the kind's HTTP status: 500 for the zero Kind.
func (k Kind) Status() int {
	if int(k) < len(kinds) {
		return kinds[k].status
	}
	return 500
}

// String returns the kind's name in snake case, such as "not_found", and
// "unspecified" for the zero Kind.
func (k Kind) String() string {
	if int(k) < len(kinds) {
		return kinds[k].name
	}
	return "Kind(" + strconv.Itoa(int(k)) + ")"
}

func (Kind) option() {}

// Option configures a declaration. A Kind is an Option; there are no others.
type Option interface{ option() }
