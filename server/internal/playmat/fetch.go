package playmat

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

const (
	// FetchTimeout bounds the whole download: dial, TLS, redirects and
	// the body.
	FetchTimeout = 10 * time.Second

	// MaxRedirects is how many redirects a fetch follows.
	MaxRedirects = 3

	maxURLLen = 2048
)

// ErrFetch is wrapped by every refusal of a pasted URL. The message is
// shown to the person who pasted it, so it says what to do and never
// echoes what the resolver or the remote host said.
var ErrFetch = errors.New("could not fetch that link")

func fetchErr(format string, a ...any) error {
	return fmt.Errorf("%w: %s", ErrFetch, fmt.Sprintf(format, a...))
}

// Fetcher downloads one image from a URL a person pasted. The server is
// the only party that contacts the remote host: the bytes are stored
// like an upload, and the other players' browsers never see the URL.
//
// A Fetcher is the SSRF guard (ADR 0128 §3):
//
//   - https only.
//   - The address is checked where it is dialled, on the IP that was
//     resolved, not on the host name in the URL. A name that resolves to
//     127.0.0.1, or that resolves to a public address when it is checked
//     and a private one when it is dialled (DNS rebinding), cannot get
//     through, because the check is the dial.
//   - Every redirect is a new request through the same dialer, so a
//     redirect to a private address is refused the same way. At most
//     MaxRedirects.
//   - No proxy from the environment, no cookie jar, no credentials, no
//     Referer. FetchTimeout for everything, MaxUploadBytes for the body.
//
// The zero value is not usable; use NewFetcher.
type Fetcher struct {
	client  *http.Client
	timeout time.Duration
}

// fetcherHooks are the test seams. Production leaves them nil.
type fetcherHooks struct {
	// lookup replaces DNS. It returns the addresses for host.
	lookup func(ctx context.Context, host string) ([]netip.Addr, error)
	// allow replaces the address policy, for the success-path tests
	// that run against a server on 127.0.0.1.
	allow func(netip.Addr) bool
	// tlsConfig replaces the transport's TLS configuration, so a test
	// can trust an httptest certificate.
	tlsConfig *tls.Config
	// timeout overrides FetchTimeout for the HTTP client. Zero means use FetchTimeout.
	timeout time.Duration
}

// NewFetcher returns the production Fetcher.
func NewFetcher() *Fetcher { return newFetcher(fetcherHooks{}) }

func newFetcher(h fetcherHooks) *Fetcher {
	allow := h.allow
	if allow == nil {
		allow = AddrAllowed
	}
	lookup := h.lookup
	if lookup == nil {
		lookup = func(ctx context.Context, host string) ([]netip.Addr, error) {
			return net.DefaultResolver.LookupNetIP(ctx, "ip", host)
		}
	}
	control := func(_, address string, _ syscall.RawConn) error {
		ap, err := netip.ParseAddrPort(address)
		if err != nil || !allow(ap.Addr()) {
			return errors.New("address not allowed")
		}
		return nil
	}
	dialer := &net.Dialer{Timeout: FetchTimeout, Control: control}
	tr := &http.Transport{
		Proxy:                 nil,
		TLSClientConfig:       h.tlsConfig,
		TLSHandshakeTimeout:   FetchTimeout,
		ResponseHeaderTimeout: FetchTimeout,
		DisableKeepAlives:     true,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
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
				return nil, errors.New("no address")
			}
			// Refuse the host when ANY of its addresses is not allowed,
			// rather than skipping to the public one: a name that
			// answers with a private address is not one to be talked
			// into a retry on.
			var last error
			for _, ip := range ips {
				// Dial the address that was checked, not the name, so
				// a second lookup cannot return something else. The
				// Control hook checks it once more at connect time.
				c, derr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
				if derr == nil {
					return c, nil
				}
				last = derr
			}
			return nil, last
		},
	}
	timeout := h.timeout
	if timeout == 0 {
		timeout = FetchTimeout
	}
	return &Fetcher{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
			// No Jar: no cookies are stored or sent.
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) > MaxRedirects {
					return fetchErr("too many redirects")
				}
				if req.URL.Scheme != "https" {
					return fetchErr("the link redirects to a page that is not https")
				}
				// net/http adds the previous URL as a Referer on a redirect.
				// It would tell the next host which link was pasted.
				req.Header.Del("Referer")
				return nil
			},
		},
		timeout: timeout,
	}
}

// ParseURL validates the URL a person pasted before any network is
// touched.
func ParseURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fetchErr("paste a link to an image")
	}
	if len(raw) > maxURLLen {
		return nil, fetchErr("that link is too long")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil, fetchErr("that is not a web address")
	}
	if u.Scheme != "https" {
		return nil, fetchErr("only https links are accepted")
	}
	if u.User != nil {
		return nil, fetchErr("links with a user name or password are not accepted")
	}
	return u, nil
}

// Fetch downloads the image at rawURL and returns its bytes, which the
// caller passes to Normalize. It checks nothing about the bytes: the
// Content-Type the remote host sent means nothing, and Normalize is the
// one place that decides what an image is.
func (f *Fetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	u, err := ParseURL(rawURL)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, f.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fetchErr("that is not a web address")
	}
	req.Header.Set("User-Agent", "cmd-and-ctrl-playmat/1")
	req.Header.Set("Accept", "image/png,image/jpeg,image/webp")
	resp, err := f.client.Do(req)
	if err != nil {
		if errors.Is(err, ErrFetch) {
			return nil, unwrapFetchErr(err)
		}
		// Never echo the transport's error: it can name the address
		// the guard refused, which is a probe of the server's network.
		return nil, fetchErr("the server could not reach that address")
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fetchErr("that address answered with status %d", resp.StatusCode)
	}
	if resp.ContentLength > MaxUploadBytes {
		return nil, ErrTooLarge
	}
	data, err := ReadLimited(resp.Body)
	if err != nil {
		if errors.Is(err, ErrTooLarge) {
			return nil, err
		}
		return nil, fetchErr("the download did not finish")
	}
	return data, nil
}

// unwrapFetchErr returns the ErrFetch-wrapped error a CheckRedirect
// produced, which http.Client wraps in a *url.Error.
func unwrapFetchErr(err error) error {
	var ue *url.Error
	if errors.As(err, &ue) && errors.Is(ue.Err, ErrFetch) {
		return ue.Err
	}
	return err
}

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

// AddrAllowed reports whether the server may connect to ip on behalf
// of a person who pasted a link. It is a deny list of every address
// that is not a public unicast address: loopback, private (RFC 1918
// and IPv6 unique-local), link-local (which includes 169.254.169.254,
// the cloud metadata address), carrier-grade NAT, multicast,
// unspecified, and the reserved and documentation ranges.
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
