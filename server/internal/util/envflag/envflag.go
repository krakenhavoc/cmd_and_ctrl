// Package envflag interprets boolean-ish environment variables
// consistently across the server's dev / ops knobs
// (CMDCTRL_TRUST_FORWARDED, CMDCTRL_DEV_RELAX_RATE_LIMITS,
// CMDCTRL_SECURE_COOKIES, ...).
package envflag

import "strings"

// Truthy reports whether val reads as an affirmative boolean. The
// empty string and the common negative spellings ("0", "false",
// "no", "off", case-insensitive) are falsy; any other non-empty
// value is truthy — matching the shell convention where setting the
// variable at all usually means "on".
func Truthy(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "", "0", "false", "no", "off":
		return false
	}
	return true
}
