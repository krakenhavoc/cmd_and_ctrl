// Package clientip names the client a request came from. Every per-IP
// rate limiter and every log line that records a remote address asks
// FromRequest, so there is one answer to "who is this" (#3334).
//
// The hosts sit behind a chain of proxies:
//
//	browser -> Cloudflare edge -> cloudflared (on the VM) -> Caddy :80 -> 127.0.0.1:8080
//
// so the socket peer, r.RemoteAddr, is always Caddy on loopback. With
// CMDCTRL_TRUST_FORWARDED set, the client is read from the headers that
// chain adds, in this order:
//
//  1. CF-Connecting-IP. Cloudflare sets it at the edge to the address
//     that connected to it, overwriting any value the client sent, and
//     cloudflared and Caddy pass it through unchanged.
//  2. X-Forwarded-For, read from the RIGHT, skipping loopback entries.
//     Cloudflare appends the client's address to whatever the client
//     sent, and Caddy (trusting its loopback peer, deploy/Caddyfile)
//     appends cloudflared's loopback address after it. The rightmost
//     entry that is not loopback is therefore the one Cloudflare wrote;
//     everything left of it arrived in the client's own header and is
//     never read, so a caller cannot mint a fresh bucket per request.
//  3. The host part of r.RemoteAddr.
//
// Without the flag only r.RemoteAddr is used, because a header is then
// whatever the caller chose to send. The flag is safe only when the
// origin is reachable through the tunnel alone: anyone who can reach
// Caddy or the server directly can set CF-Connecting-IP to anything.
package clientip

import (
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/util/envflag"
)

// EnvTrustForwarded is the variable that turns header parsing on. CD
// sets it on both hosts ("Sync server env (proxy and cookie flags)").
const EnvTrustForwarded = "CMDCTRL_TRUST_FORWARDED"

// Trusted reports whether the forwarding headers are believed. It reads
// the environment on every call, like the other CMDCTRL runtime knobs,
// so flipping it does not need a restart.
func Trusted() bool {
	return envflag.Truthy(os.Getenv(EnvTrustForwarded))
}

// FromRequest returns the client address for r: a bare IP with no port
// and no zone, IPv4-mapped IPv6 folded to IPv4, so one client always
// gets one string. If RemoteAddr is not an address at all it is
// returned as it is, which keeps tests with odd RemoteAddrs keyed.
func FromRequest(r *http.Request) string {
	return resolve(r, Trusted())
}

func resolve(r *http.Request, trust bool) string {
	if trust {
		if ip, ok := parse(r.Header.Get("CF-Connecting-IP")); ok {
			return ip
		}
		if ip, ok := fromForwardedFor(r.Header.Values("X-Forwarded-For")); ok {
			return ip
		}
	}
	return remoteHost(r.RemoteAddr)
}

// fromForwardedFor walks the X-Forwarded-For entries from the right,
// over every header line in order, and returns the first one that is
// not loopback. An entry it cannot parse stops the walk: it cannot
// have been written by a hop we trust, so nothing left of it is either.
func fromForwardedFor(lines []string) (string, bool) {
	entries := strings.Split(strings.Join(lines, ","), ",")
	for i := len(entries) - 1; i >= 0; i-- {
		addr, ok := parseAddr(entries[i])
		if !ok {
			return "", false
		}
		if addr.IsLoopback() {
			continue
		}
		return addr.String(), true
	}
	return "", false
}

func remoteHost(remote string) string {
	host, _, err := net.SplitHostPort(remote)
	if err != nil {
		host = remote
	}
	if ip, ok := parse(host); ok {
		return ip
	}
	if host == "" {
		return remote
	}
	return host
}

func parse(s string) (string, bool) {
	addr, ok := parseAddr(s)
	if !ok {
		return "", false
	}
	return addr.String(), true
}

// parseAddr accepts a bare IPv4 or IPv6 address, a bracketed IPv6
// address, or either with a port.
func parseAddr(s string) (netip.Addr, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return netip.Addr{}, false
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		s = host
	}
	s = strings.TrimSuffix(strings.TrimPrefix(s, "["), "]")
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap().WithZone(""), true
}
