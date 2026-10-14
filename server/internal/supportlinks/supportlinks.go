// Package supportlinks reads the operator's optional supporter links
// (#3558, ADR 0179 "Supporter links"): a Ko-fi page and a Patreon page,
// each from one environment variable. They are a voluntary thank-you jar
// shown on the funding page; nothing in the game depends on them.
//
// The server never follows a link. It checks each value once, at boot,
// and serves the valid set on GET /config. Unset, the default, means
// the link is not shown.
//
// A value is an absolute https URL on that service's own domain (the
// bare domain or www), with no credentials, so a typo or a pasted
// tracker cannot send players somewhere else.
package supportlinks

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// MaxBytes bounds a link. A real page URL is well under 200 bytes.
const MaxBytes = 512

// Env names of the variables.
const (
	KofiEnv    = "CMDCTRL_SUPPORT_KOFI"
	PatreonEnv = "CMDCTRL_SUPPORT_PATREON"
)

// Links is the configured set, as GET /config serves it. A blank field
// is a link that is not shown.
type Links struct {
	Kofi    string `json:"kofi,omitempty"`
	Patreon string `json:"patreon,omitempty"`
}

// FromEnv reads both variables. Unset or blank leaves that link out. An
// invalid value is an error naming every offending variable: main stops
// the boot on it.
func FromEnv(getenv func(string) string) (Links, error) {
	var l Links
	var errs []error
	for _, s := range []struct {
		env    string
		domain string
		field  *string
	}{
		{KofiEnv, "ko-fi.com", &l.Kofi},
		{PatreonEnv, "patreon.com", &l.Patreon},
	} {
		raw := strings.TrimSpace(getenv(s.env))
		if raw == "" {
			continue
		}
		if err := Validate(raw, s.domain); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", s.env, err))
			continue
		}
		*s.field = raw
	}
	if len(errs) > 0 {
		return Links{}, errors.Join(errs...)
	}
	return l, nil
}

// Validate checks that raw is an https URL whose host is domain or
// www.domain. The client applies the same rule and drops a link that
// fails it.
func Validate(raw, domain string) error {
	if len(raw) > MaxBytes {
		return fmt.Errorf("a support link is at most %d bytes", MaxBytes)
	}
	for _, c := range raw {
		if c <= ' ' || c >= 0x7f || strings.ContainsRune("\"'`\\<>{}|^", c) {
			return errors.New("a support link is plain ASCII, with no spaces, quotes, backslashes, braces or angle brackets")
		}
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("a support link must be an https URL: %v", err)
	}
	if u.Scheme != "https" || u.Host == "" || u.Opaque != "" {
		return fmt.Errorf("a support link must be an https URL, such as https://%s/yourname", domain)
	}
	if u.User != nil {
		return errors.New("a support link must not hold a user name or password")
	}
	if u.Port() != "" {
		return errors.New("a support link must not name a port")
	}
	if h := strings.ToLower(u.Hostname()); h != domain && h != "www."+domain {
		return fmt.Errorf("a support link must be on %s (or www.%s), not %q", domain, domain, u.Hostname())
	}
	return nil
}
