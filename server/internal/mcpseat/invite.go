package mcpseat

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

// DefaultOrigins are the servers `join` accepts when the owner names none
// (§8): the two labxp hosts, and localhost on any port.
var DefaultOrigins = []string{
	"https://cmd.labxp.io",
	"https://cmd-dev.labxp.io",
	"http://localhost:*",
	"http://127.0.0.1:*",
}

// OriginAllowlist is the set of servers this seat will talk to. It is set
// from the owner's MCP configuration (--allow-origin), never by the model:
// an invite URL from a hostile chat line cannot point the seat at a
// server that feeds it a forged table.
type OriginAllowlist struct {
	entries []originPattern
}

type originPattern struct {
	scheme, host string
	port         string // "" = the scheme's default, "*" = any
}

// NewOriginAllowlist parses entries of the form scheme://host[:port],
// where the port may be "*". An empty list means DefaultOrigins.
func NewOriginAllowlist(entries []string) (*OriginAllowlist, error) {
	if len(entries) == 0 {
		entries = DefaultOrigins
	}
	al := &OriginAllowlist{}
	for _, e := range entries {
		p, err := parseOriginPattern(e)
		if err != nil {
			return nil, err
		}
		al.entries = append(al.entries, p)
	}
	return al, nil
}

func parseOriginPattern(s string) (originPattern, error) {
	s = strings.TrimRight(strings.TrimSpace(s), "/")
	scheme, rest, ok := strings.Cut(s, "://")
	if !ok || (scheme != "http" && scheme != "https") || rest == "" || strings.ContainsAny(rest, "/?#@") {
		return originPattern{}, fmt.Errorf("--allow-origin %q: want http(s)://host[:port]", s)
	}
	host, port := rest, ""
	if h, p, err := net.SplitHostPort(rest); err == nil {
		host, port = h, p
	}
	if host == "" {
		return originPattern{}, fmt.Errorf("--allow-origin %q: no host", s)
	}
	return originPattern{scheme: scheme, host: strings.ToLower(host), port: port}, nil
}

// Allows reports whether u's origin is on the list.
func (al *OriginAllowlist) Allows(u *url.URL) bool {
	host := strings.ToLower(u.Hostname())
	port := u.Port()
	for _, p := range al.entries {
		if p.scheme != u.Scheme || p.host != host {
			continue
		}
		if p.port == "*" || p.port == port {
			return true
		}
		if p.port == "" && port == "" {
			return true
		}
	}
	return false
}

// String lists the entries, for an error message.
func (al *OriginAllowlist) String() string {
	out := make([]string, 0, len(al.entries))
	for _, p := range al.entries {
		s := p.scheme + "://" + p.host
		if p.port != "" {
			s += ":" + p.port
		}
		out = append(out, s)
	}
	return strings.Join(out, ", ")
}

// inviteKind is what a pasted link asks for.
type inviteKind int

const (
	inviteJoin inviteKind = iota
	inviteReclaim
)

// invite is a parsed invite or reclaim link.
type invite struct {
	Origin string // scheme://host[:port], no trailing slash
	GameID uuid.UUID
	Token  string // the ?t= value: the invite token or the reclaim ticket
	Kind   inviteKind
}

// errNotInvite is returned for anything that is not a link this seat
// understands.
var errNotInvite = errors.New("that is not an invite link: want <origin>/#/games/<id>/join?t=<token> (or a /reclaim?t=<ticket> link)")

// parseInvite reads the links the client and the Discord bot hand out:
//
//	<origin>/#/games/<game id>/join?t=<invite token>
//	<origin>/#/games/<game id>/reclaim?t=<reclaim ticket>
//
// The fragment carries the route, so the query is the fragment's own.
func parseInvite(raw string) (invite, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return invite{}, errNotInvite
	}
	route := u.Fragment
	if route == "" {
		// A link with the route in the path rather than the fragment.
		route = u.Path
		if u.RawQuery != "" {
			route += "?" + u.RawQuery
		}
	}
	route = strings.TrimPrefix(route, "/")
	path, query, _ := strings.Cut(route, "?")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 3 || parts[0] != "games" {
		return invite{}, errNotInvite
	}
	id, err := uuid.Parse(parts[1])
	if err != nil {
		return invite{}, errNotInvite
	}
	var kind inviteKind
	switch parts[2] {
	case "join":
		kind = inviteJoin
	case "reclaim":
		kind = inviteReclaim
	default:
		return invite{}, errNotInvite
	}
	q, err := url.ParseQuery(query)
	if err != nil || q.Get("t") == "" {
		return invite{}, errNotInvite
	}
	return invite{
		Origin: u.Scheme + "://" + u.Host,
		GameID: id,
		Token:  q.Get("t"),
		Kind:   kind,
	}, nil
}
