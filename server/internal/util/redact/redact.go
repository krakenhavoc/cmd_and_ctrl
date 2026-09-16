// Package redact strips credentials out of text before it is published
// anywhere a session token must not appear — today, the body of an
// in-app bug report filed as a GitHub issue (#721, ADR 0017 §9).
//
// Client-supplied text is the risk. The browser's log once carried the
// WebSocket URL verbatim, `?token=` and all, and the bug-report renderer
// inlined it, so every report with a connect line published the
// reporter's session token. The client now redacts at the source
// (client/src/lib/redact.ts); this package is the server's copy of the
// same rules, so an old tab, a curl, or a log line nobody thought about
// cannot leak through the server either. Keep the two rule sets in step.
package redact

import "regexp"

// Redacted replaces a credential value. Keys are kept, so a triager can
// still see that a token was sent without being able to use it.
const Redacted = "REDACTED"

const (
	// secretKey matches a key that names a credential: token,
	// access_token, refresh_token, id_token, client_secret, password,
	// the reclaim ticket, … A suffix match rather than a list, so a new
	// `*_token` is covered without anyone remembering to add it.
	secretKey = `[a-z0-9_-]*(?:token|secret|password|passwd|ticket)`

	// value is one key's value: everything up to the next parameter or
	// fragment separator, whitespace, quote or bracket. Percent-escapes
	// are part of the value EXCEPT %23 (#) and %26 (&) — and their
	// double-encoded %2523 / %2526 — which end it. That keeps a
	// URL-encoded URL (`…%3Ftoken%3Dabc%26game%3D1`) from losing its
	// trailing parameters to the redaction.
	//
	// The escape branches: a single escape other than %23/%25/%26; a
	// double escape (%25XX) other than %2523/%2526.
	value = `(?:[^&#\s"'<>()\[\]{},;%\\]|%(?:[013-9a-f][0-9a-f]|2[0-2479a-f])|%25(?:[013-9a-f][0-9a-f]|2[0-2457-9a-f]))+`

	// eq is a literal, percent-encoded or double-encoded equals sign
	// (`=`, `%3D`, `%253D` — the last is a URL carried inside an encoded
	// parameter of another URL).
	eq = `(=|%3d|%253d)`
)

var (
	// key=value anywhere, the key not glued to a preceding word
	// character ("notatoken=x" is left alone).
	keyValueRE = regexp.MustCompile(`(?i)(^|[^a-z0-9_-])(` + secretKey + `)` + eq + value)

	// `t` is the invite and reclaim token (`#/games/<id>/join?t=…`). Too
	// short to trust in prose ("at t=3"), so only redacted as a query
	// parameter: after ?, &, &amp;, or their encoded and double-encoded
	// forms.
	queryOnlyRE = regexp.MustCompile(`(?i)(\?|&amp;|&|%253f|%2526|%3f|%26)(t)` + eq + value)

	// JSON-shaped: "token":"abc", and the escaped form a stringified
	// object picks up inside another string (\"token\":\"abc\").
	jsonRE = regexp.MustCompile(`(?i)(\\?")(` + secretKey + `)(\\?"\s*:\s*\\?")(?:[^"\\]|\\[^"])*(\\?")`)

	// `Authorization: Bearer <credential>`. Case-sensitive and
	// length-gated on purpose: "Bearer of the Heavens" is a card name,
	// and a real credential is never four letters long.
	bearerRE = regexp.MustCompile(`(^|[^A-Za-z])(Bearer)(\s+|%20|\+)[A-Za-z0-9._~+/=-]{16,}`)
)

// Secrets returns s with every credential value it recognises replaced
// by Redacted: token-bearing query parameters (literal and URL-encoded),
// the invite `t`, JSON token fields, and Bearer credentials. Idempotent.
func Secrets(s string) string {
	if s == "" {
		return s
	}
	s = keyValueRE.ReplaceAllString(s, "${1}${2}${3}"+Redacted)
	s = queryOnlyRE.ReplaceAllString(s, "${1}${2}${3}"+Redacted)
	s = jsonRE.ReplaceAllString(s, "${1}${2}${3}"+Redacted+"${4}")
	s = bearerRE.ReplaceAllString(s, "${1}${2}${3}"+Redacted)
	return s
}
