package tenant

import (
	"fmt"
	"strings"
)

// Mode selects how Resolve determines the tenant of a request.
type Mode int

const (
	// B2B requires every request to name its tenant through the
	// X-Tenant-ID header. It is the zero value, so an unset Mode is the
	// stricter of the two.
	B2B Mode = iota

	// B2C attributes every request to one configured tenant and
	// rejects requests that name a tenant themselves.
	B2C
)

// ParseMode maps an operator-provided string to a Mode. It accepts "B2B" and
// "B2C" and returns an error for anything else, so that a misconfigured mode
// fails at startup rather than silently defaulting.
func ParseMode(mode string) (Mode, error) {
	mode = strings.ToUpper(mode)

	switch mode {
	case "B2B":
		return B2B, nil
	case "B2C":
		return B2C, nil
	default:
		return B2B, fmt.Errorf("invalid mode %q", mode)
	}
}
