// Package shoplinks reads the operator's shop link configuration (ADR
// 0179): one environment variable per retailer, each a URL template
// holding the placeholder {url}. The client puts the percent-encoded
// address of a card search or a pasted-list cart at the retailer in
// place of {url}, so a retailer's referral wrapper or parameter can be
// set without a code change.
//
// The server never builds a link and never follows one. It checks each
// template once, at boot, and serves the valid set on GET /config. With
// no variable set, which is the default, the set is empty and the client
// renders no shop link anywhere.
//
// Two shapes are accepted:
//
//   - A wrapper: an absolute https URL with {url} once, after the host,
//     for example https://partner.example/c/123?u={url}. The client puts
//     the destination there percent-encoded.
//   - A direct link: {url} first, optionally followed by &name=value
//     pairs, for example {url}&partner=cmdctrl. The client uses the
//     destination as it is (every destination already carries a query)
//     and appends the pairs.
package shoplinks

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// Placeholder is what a template holds where the destination goes.
const Placeholder = "{url}"

// MaxTemplateBytes bounds a template. A real wrapper is under 200 bytes.
const MaxTemplateBytes = 1024

// Retailer is one shop the client knows how to build links for. The
// client holds the matching destination formats (lib/shopLinks.ts);
// adding a retailer here without one there shows nothing.
type Retailer struct {
	// ID is the wire name, the key the client builds links by.
	ID string
	// Name is what the link says.
	Name string
	// Env is the variable that holds the template.
	Env string
}

// Retailers is every supported shop, in the order the links render.
var Retailers = []Retailer{
	{ID: "tcgplayer", Name: "TCGplayer", Env: "CMDCTRL_SHOP_TCGPLAYER"},
	{ID: "cardkingdom", Name: "Card Kingdom", Env: "CMDCTRL_SHOP_CARDKINGDOM"},
}

// Shop is one configured retailer, as GET /config serves it.
type Shop struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Template string `json:"template"`
}

// FromEnv reads every retailer's variable. Unset or blank leaves that
// retailer out. A value that ValidateTemplate refuses is an error naming
// the variable: main stops the boot on it, so a typo is found at deploy
// rather than as a broken link. The result is never nil.
func FromEnv(getenv func(string) string) ([]Shop, error) {
	shops := []Shop{}
	var errs []error
	for _, r := range Retailers {
		raw := strings.TrimSpace(getenv(r.Env))
		if raw == "" {
			continue
		}
		if err := ValidateTemplate(raw); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", r.Env, err))
			continue
		}
		shops = append(shops, Shop{ID: r.ID, Name: r.Name, Template: raw})
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return shops, nil
}

// ValidateTemplate checks one template against the two shapes in the
// package comment. The client applies the same rules (parseShops) and
// drops a template that fails them, so a server and client that ever
// disagree fail closed.
func ValidateTemplate(raw string) error {
	if len(raw) > MaxTemplateBytes {
		return fmt.Errorf("a shop link template is at most %d bytes", MaxTemplateBytes)
	}
	if n := strings.Count(raw, Placeholder); n != 1 {
		return fmt.Errorf("a shop link template holds %s exactly once (found %d)", Placeholder, n)
	}
	for _, c := range strings.Replace(raw, Placeholder, "", 1) {
		if c <= ' ' || c >= 0x7f || strings.ContainsRune("\"'`\\<>{}|^", c) {
			return errors.New("a shop link template is plain ASCII, with no spaces, quotes, backslashes, braces or angle brackets")
		}
	}
	if strings.HasPrefix(raw, Placeholder) {
		rest := raw[len(Placeholder):]
		if rest == "" {
			return nil
		}
		if !strings.HasPrefix(rest, "&") || len(rest) == 1 {
			return fmt.Errorf("after a leading %s, a template may only add &name=value pairs", Placeholder)
		}
		for _, pair := range strings.Split(rest[1:], "&") {
			name, _, _ := strings.Cut(pair, "=")
			if name == "" {
				return fmt.Errorf("after a leading %s, every pair needs a name", Placeholder)
			}
		}
		if _, err := url.ParseQuery(rest[1:]); err != nil {
			return fmt.Errorf("after a leading %s, the pairs do not parse: %v", Placeholder, err)
		}
		return nil
	}
	u, err := url.Parse(strings.Replace(raw, Placeholder, "x", 1))
	if err != nil {
		return fmt.Errorf("a shop link template must be an https URL: %v", err)
	}
	if u.Scheme != "https" || u.Host == "" || u.User != nil || u.Opaque != "" {
		return fmt.Errorf("a shop link template must be an https URL with a host, such as https://partner.example/c/123?u=%s, or start with %s", Placeholder, Placeholder)
	}
	// The placeholder must come after the host, so a destination can
	// never choose where the link goes.
	hostEnd := strings.Index(raw, "://") + 3 + len(u.Host)
	if idx := strings.Index(raw, Placeholder); idx < hostEnd {
		return fmt.Errorf("%s must come after the host", Placeholder)
	}
	return nil
}
