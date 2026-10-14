package modelkeys

import (
	"errors"
	"net"
	"net/netip"
	"net/url"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/netguard"
)

// ErrInvalid wraps every refusal of what a person typed. The text after
// it is shown to them and never echoes the key.
var ErrInvalid = errors.New("invalid")

// InvalidError is a refusal with a sentence for the person.
type InvalidError struct{ Msg string }

func (e *InvalidError) Error() string { return e.Msg }
func (e *InvalidError) Unwrap() error { return ErrInvalid }

func invalid(msg string) error { return &InvalidError{Msg: msg} }

// Limits on what a person may type.
const (
	MinKeyLen   = 8
	MaxKeyLen   = 512
	MaxModelLen = 128
	MaxURLLen   = 512

	MinGamesPerDay     = 1
	MaxGamesPerDay     = 100
	DefaultGamesPerDay = 10

	MinCallsPerGame     = 1
	MaxCallsPerGame     = 5000
	DefaultCallsPerGame = 500
)

// ValidateKey checks an API key's shape: printable ASCII with no spaces,
// within the length limits. It says nothing about whether the provider
// will take it; that is what Test is for.
func ValidateKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if len(key) < MinKeyLen || len(key) > MaxKeyLen {
		return "", invalid("paste the whole API key: it should be between 8 and 512 characters")
	}
	for i := 0; i < len(key); i++ {
		if c := key[i]; c < 0x21 || c > 0x7e {
			return "", invalid("an API key has no spaces or special characters; paste it again")
		}
	}
	return key, nil
}

// ValidateModel checks a model id: the characters providers use in
// theirs (letters, digits and . _ - : / @ +), within the length limit.
func ValidateModel(m string) (string, error) {
	m = strings.TrimSpace(m)
	if m == "" {
		return "", invalid("name the model your seats should use")
	}
	if len(m) > MaxModelLen {
		return "", invalid("that model name is too long")
	}
	for _, c := range m {
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case strings.ContainsRune("._-:/@+", c):
		default:
			return "", invalid("a model name has only letters, digits and . _ - : / @ +")
		}
	}
	return m, nil
}

// ValidateLimits checks the owner's two limits.
func ValidateLimits(gamesPerDay, callsPerGame int) error {
	if gamesPerDay < MinGamesPerDay || gamesPerDay > MaxGamesPerDay {
		return invalid("games a day must be between 1 and 100")
	}
	if callsPerGame < MinCallsPerGame || callsPerGame > MaxCallsPerGame {
		return invalid("calls per game must be between 1 and 5000")
	}
	return nil
}

// ValidateCustomURL checks a "custom" provider's base URL before it is
// saved, and returns it normalised (no trailing slash).
//
// These are the save-time checks, for a clear message: https, a host
// name, port 443, no user name or password, no query or fragment, and
// not an address literal or a name that is plainly local. They are NOT
// the guard. The guard is netguard's dial-time check, which runs on
// every request whatever this said, because a name that is public today
// can resolve to 127.0.0.1 tomorrow.
func ValidateCustomURL(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", invalid("enter the provider's API address, starting https://")
	}
	if len(raw) > MaxURLLen {
		return "", invalid("that address is too long")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.Opaque != "" {
		return "", invalid("that is not a web address")
	}
	if u.Scheme != "https" {
		return "", invalid("only https addresses are accepted")
	}
	if u.User != nil {
		return "", invalid("an address with a user name or password is not accepted; put the key in the key field")
	}
	if u.RawQuery != "" || u.Fragment != "" || u.ForceQuery {
		return "", invalid("leave the query string off the address")
	}
	if p := u.Port(); p != "" && p != "443" {
		return "", invalid("only the standard https port (443) is accepted")
	}
	host := strings.ToLower(strings.TrimSuffix(u.Hostname(), "."))
	if ip, perr := netip.ParseAddr(host); perr == nil {
		if !netguard.AddrAllowed(ip) {
			return "", invalid("that address is on a private or local network, which this server will not contact")
		}
		return "", invalid("use the provider's host name, not an IP address")
	}
	if net.ParseIP(host) != nil || !strings.Contains(host, ".") || localName(host) {
		return "", invalid("that address is on a private or local network, which this server will not contact")
	}
	u.Host = host
	u.Path = strings.TrimRight(u.Path, "/")
	u.RawPath = ""
	return u.String(), nil
}

// localName is a host name that is local by convention. A convenience
// for the message only: the dial check refuses whatever it resolves to.
func localName(host string) bool {
	for _, suffix := range []string{"localhost", ".local", ".internal", ".lan", ".home", ".arpa", ".localdomain"} {
		if host == strings.TrimPrefix(suffix, ".") || strings.HasSuffix(host, suffix) {
			return true
		}
	}
	return false
}

// last4 is the tail the Settings page shows.
func last4(key string) string {
	if len(key) <= 4 {
		return ""
	}
	return key[len(key)-4:]
}
