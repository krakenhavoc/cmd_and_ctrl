// Package netguard is the server's guard for outbound requests to an
// address a person chose: an https client that will only connect to a
// public unicast address, checked where it is dialled.
//
// Two features send requests to a place a player named. A playmat link
// (ADR 0128 §3) fetches one image; a model key with a custom endpoint
// (ADR 0182) posts to an OpenAI-compatible server for every decision an
// AI seat makes. Either is a server-side request forgery vector into
// the host's own network (the cloud metadata address, a database on
// localhost, a router on 192.168.x) unless the address is checked, and
// checked at the right moment: a check at save time is defeated by a
// name that resolves to a public address when it is checked and a
// private one when it is dialled (DNS rebinding). So the check IS the
// dial:
//
//   - The host name is resolved once. If ANY address it answers with is
//     refused, the name is refused: a name that answers with a private
//     address is not one to be talked into a retry on.
//   - The address that passed is the address dialled (not the name), so
//     a second lookup cannot answer differently.
//   - net.Dialer.Control checks it once more at connect time, on the
//     address the socket is actually connecting to.
//
// AddrAllowed is the address policy; NewClient builds the client.
package netguard

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"syscall"
	"time"
)

var (
	// ErrAddressRefused is the dial's refusal of an address that is not
	// public unicast. Callers never show it (or the address) to a
	// person: naming what was refused is a probe of the host's network.
	ErrAddressRefused = errors.New("netguard: address not allowed")
	// ErrNotHTTPS is a request whose URL is not https.
	ErrNotHTTPS = errors.New("netguard: only https is allowed")
	// ErrResponseTooLarge is a response body over Options.MaxResponseBytes.
	ErrResponseTooLarge = errors.New("netguard: response too large")
)

// Options configure NewClient. The zero value is valid and strict:
// production passes only Timeout and MaxResponseBytes.
type Options struct {
	// Timeout bounds the whole request: dial, TLS, headers and body.
	// Zero means 10 seconds.
	Timeout time.Duration
	// MaxResponseBytes caps the response body. Reading past it fails
	// with ErrResponseTooLarge. Zero means 1 MiB.
	MaxResponseBytes int64

	// The rest are TEST SEAMS. Production leaves them nil.

	// Allow replaces AddrAllowed, so a success-path test can reach an
	// httptest server on 127.0.0.1 while every other address stays
	// refused.
	Allow func(netip.Addr) bool
	// Lookup replaces DNS.
	Lookup func(ctx context.Context, host string) ([]netip.Addr, error)
	// TLSConfig replaces the transport's TLS configuration, so a test
	// can trust an httptest certificate.
	TLSConfig *tls.Config
	// Port replaces the port dialled, so a test can point a URL that
	// must say 443 at an httptest server's port. The address checks are
	// unchanged.
	Port string
}

const (
	defaultTimeout  = 10 * time.Second
	defaultMaxBytes = 1 << 20
)

// NewClient returns an http.Client that:
//
//   - sends only https requests (ErrNotHTTPS otherwise),
//   - connects only to addresses Allow (AddrAllowed) accepts, checked at
//     dial time as the package comment describes,
//   - follows no redirect (a 3xx is returned to the caller as it is),
//   - uses no proxy from the environment and keeps no cookies,
//   - gives up after Timeout, and
//   - fails a body read past MaxResponseBytes.
func NewClient(o Options) *http.Client {
	timeout := o.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	maxBytes := o.MaxResponseBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxBytes
	}
	allow := o.Allow
	if allow == nil {
		allow = AddrAllowed
	}
	lookup := o.Lookup
	if lookup == nil {
		lookup = func(ctx context.Context, host string) ([]netip.Addr, error) {
			return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		}
	}
	dialer := &net.Dialer{Timeout: timeout, Control: control(allow)}
	tr := &http.Transport{
		Proxy:                  nil,
		TLSClientConfig:        o.TLSConfig,
		TLSHandshakeTimeout:    timeout,
		ResponseHeaderTimeout:  timeout,
		MaxResponseHeaderBytes: 64 << 10,
		IdleConnTimeout:        30 * time.Second,
		MaxIdleConnsPerHost:    2,
		ForceAttemptHTTP2:      true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if o.Port != "" {
				if host, _, err := net.SplitHostPort(addr); err == nil {
					addr = net.JoinHostPort(host, o.Port)
				}
			}
			return dialChecked(ctx, dialer, lookup, allow, network, addr)
		},
	}
	return &http.Client{
		Transport: &capTransport{inner: tr, max: maxBytes},
		Timeout:   timeout,
		// No Jar: no cookies are stored or sent.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			// Not followed: a redirect is a second address the first
			// one chose, and an API endpoint has no reason to send one.
			return http.ErrUseLastResponse
		},
	}
}

// control is the dialer's last word: it checks the address the socket
// is about to connect to, after every lookup is done.
func control(allow func(netip.Addr) bool) func(string, string, syscall.RawConn) error {
	return func(_, address string, _ syscall.RawConn) error {
		ap, err := netip.ParseAddrPort(address)
		if err != nil || !allow(ap.Addr()) {
			return ErrAddressRefused
		}
		return nil
	}
}

// dialChecked resolves addr's host once, refuses it when any answer is
// refused, and dials the checked addresses (not the name) in order.
func dialChecked(ctx context.Context, dialer *net.Dialer,
	lookup func(context.Context, string) ([]netip.Addr, error), allow func(netip.Addr) bool,
	network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	var ips []netip.Addr
	if ip, perr := netip.ParseAddr(host); perr == nil {
		ips = []netip.Addr{ip}
	} else {
		ips, err = lookup(ctx, host)
		if err != nil {
			return nil, err
		}
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("netguard: %s has no address", host)
	}
	for _, ip := range ips {
		if !allow(ip) {
			return nil, ErrAddressRefused
		}
	}
	var last error
	for _, ip := range ips {
		c, derr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.Unmap().String(), port))
		if derr == nil {
			return c, nil
		}
		last = derr
	}
	return nil, last
}

// capTransport refuses a non-https request before it is sent and caps
// the body of every response.
type capTransport struct {
	inner http.RoundTripper
	max   int64
}

func (t *capTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL == nil || req.URL.Scheme != "https" {
		if req.Body != nil {
			_ = req.Body.Close()
		}
		return nil, ErrNotHTTPS
	}
	resp, err := t.inner.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if resp.ContentLength > t.max {
		_ = resp.Body.Close()
		return nil, ErrResponseTooLarge
	}
	resp.Body = &capBody{rc: resp.Body, left: t.max}
	return resp, nil
}

// capBody fails a read that would pass the cap, rather than silently
// truncating: a cut-off JSON reply is a decode error that hides why.
type capBody struct {
	rc   io.ReadCloser
	left int64
}

func (b *capBody) Read(p []byte) (int, error) {
	if b.left <= 0 {
		// One byte more than the cap proves the body is over it.
		var one [1]byte
		n, err := b.rc.Read(one[:])
		if n > 0 {
			return 0, ErrResponseTooLarge
		}
		return 0, err
	}
	if int64(len(p)) > b.left {
		p = p[:b.left]
	}
	n, err := b.rc.Read(p)
	b.left -= int64(n)
	return n, err
}

func (b *capBody) Close() error { return b.rc.Close() }

// Blocked ranges beyond what netip's own predicates cover.
var blockedPrefixes = func() []netip.Prefix {
	var out []netip.Prefix
	for _, s := range []string{
		"0.0.0.0/8",       // "this network"
		"100.64.0.0/10",   // carrier-grade NAT (RFC 6598)
		"192.0.0.0/24",    // IETF protocol assignments
		"192.0.2.0/24",    // TEST-NET-1
		"198.18.0.0/15",   // benchmarking
		"198.51.100.0/24", // TEST-NET-2
		"203.0.113.0/24",  // TEST-NET-3
		"240.0.0.0/4",     // reserved, and the broadcast address
		"100::/64",        // discard-only
		"2001:db8::/32",   // documentation
	} {
		out = append(out, netip.MustParsePrefix(s))
	}
	return out
}()

var (
	nat64Prefix = netip.MustParsePrefix("64:ff9b::/96")
	sixToFour   = netip.MustParsePrefix("2002::/16")
)

// AddrAllowed reports whether the server may connect to ip on behalf of
// a person. It is a deny list of every address that is not a public
// unicast address: loopback, private (RFC 1918 and IPv6 unique-local),
// link-local (which includes 169.254.169.254, the cloud metadata
// address), carrier-grade NAT, multicast, unspecified, and the reserved
// and documentation ranges.
//
// An IPv4 address written as IPv6 (::ffff:127.0.0.1), through NAT64
// (64:ff9b::7f00:1) or 6to4 (2002:7f00:1::) is judged by the IPv4
// address inside it, so the IPv6 spelling is not a way round.
func AddrAllowed(ip netip.Addr) bool {
	if !ip.IsValid() {
		return false
	}
	// A zone ("fe80::1%eth0") has no meaning for a public host.
	if ip.Zone() != "" {
		return false
	}
	ip = ip.Unmap()
	if ip.Is6() {
		b := ip.As16()
		switch {
		case nat64Prefix.Contains(ip):
			return AddrAllowed(netip.AddrFrom4([4]byte{b[12], b[13], b[14], b[15]}))
		case sixToFour.Contains(ip):
			return AddrAllowed(netip.AddrFrom4([4]byte{b[2], b[3], b[4], b[5]}))
		}
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() {
		return false
	}
	for _, p := range blockedPrefixes {
		if p.Contains(ip) {
			return false
		}
	}
	return true
}
