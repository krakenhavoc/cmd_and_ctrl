package netguard

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// One address from every refused class, by name, so a failure says
// which class let something through.
var refusedClasses = map[string][]string{
	"loopback":         {"127.0.0.1", "127.8.9.10", "::1"},
	"private":          {"10.0.0.1", "10.255.255.255", "172.16.0.1", "172.31.255.255", "192.168.1.1", "192.168.0.10"},
	"link-local":       {"169.254.169.254", "169.254.0.1", "fe80::1", "fe80::a00:27ff:fe4e:66a1"},
	"cgnat":            {"100.64.0.1", "100.127.255.255"},
	"unique-local":     {"fc00::1", "fd12:3456:789a::1"},
	"unspecified":      {"0.0.0.0", "::", "0.1.2.3"},
	"multicast":        {"224.0.0.1", "239.255.255.250", "ff02::1", "ff0e::1"},
	"reserved":         {"255.255.255.255", "240.0.0.1", "192.0.2.1", "198.51.100.1", "203.0.113.1", "198.18.0.1", "2001:db8::1"},
	"v4 spelled as v6": {"::ffff:127.0.0.1", "::ffff:10.0.0.1", "::ffff:192.168.1.1", "::ffff:169.254.169.254"},
	"nat64 and 6to4":   {"64:ff9b::7f00:1", "64:ff9b::a9fe:a9fe", "2002:7f00:1::1", "2002:c0a8:101::1"},
	"zoned link-local": {"fe80::1%eth0"},
	"cgnat as v6":      {"::ffff:100.64.0.1"},
}

func TestAddrAllowedRefusesEveryNonPublicClass(t *testing.T) {
	for class, addrs := range refusedClasses {
		for _, s := range addrs {
			ip, err := netip.ParseAddr(s)
			if err != nil {
				t.Fatalf("parse %q: %v", s, err)
			}
			if AddrAllowed(ip) {
				t.Errorf("%s: %s is allowed, want refused", class, s)
			}
		}
	}
	for _, s := range []string{
		"1.1.1.1", "8.8.8.8", "93.184.216.34", "172.15.0.1", "172.32.0.1", "100.63.255.255",
		"100.128.0.1", "2606:4700:4700::1111", "2001:4860:4860::8888", "::ffff:8.8.8.8",
		"64:ff9b::808:808", "2002:808:808::1",
	} {
		if !AddrAllowed(netip.MustParseAddr(s)) {
			t.Errorf("%s is refused, want allowed", s)
		}
	}
	if AddrAllowed(netip.Addr{}) {
		t.Error("the zero Addr is allowed")
	}
}

// exempt is the production policy plus the httptest server's own
// address, so a success-path test reaches the server it started and
// every other loopback address is still refused.
func exempt(ip netip.Addr) bool {
	return ip == netip.MustParseAddr("127.0.0.1") || AddrAllowed(ip)
}

func tlsFor(srv *httptest.Server) *tls.Config {
	c := srv.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
	c.MinVersion = tls.VersionTLS12
	return c
}

func counting(t *testing.T, hits *atomic.Int32, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		if h != nil {
			h(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func get(t *testing.T, c *http.Client, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return c.Do(req)
}

// A host name that resolves to an address of any refused class is
// refused before a socket is opened — the DNS-rebinding case, and the
// one the issue names: a name that resolves to 127.0.0.1 or 192.168.x.
func TestANameThatResolvesToARefusedAddressIsRefusedAtDial(t *testing.T) {
	var hits atomic.Int32
	srv := counting(t, &hits, nil)
	port := strings.TrimPrefix(srv.URL, "https://127.0.0.1:")
	for class, addrs := range refusedClasses {
		for _, a := range addrs {
			ip, err := netip.ParseAddr(a)
			if err != nil {
				t.Fatal(err)
			}
			c := NewClient(Options{
				TLSConfig: tlsFor(srv),
				Lookup:    func(context.Context, string) ([]netip.Addr, error) { return []netip.Addr{ip}, nil },
			})
			_, err = get(t, c, "https://rebind.example:"+port+"/")
			if !errors.Is(err, ErrAddressRefused) {
				t.Errorf("%s: %s: err = %v, want ErrAddressRefused", class, a, err)
			}
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("the server was reached %d times; the guard must refuse before connecting", n)
	}
}

// One refused answer refuses the whole name, in either order, without
// dialling the public one first.
func TestANameWithOneRefusedAnswerIsRefused(t *testing.T) {
	for name, addrs := range map[string][]string{
		"public then loopback": {"8.8.8.8", "127.0.0.1"},
		"private then public":  {"192.168.1.5", "8.8.8.8"},
	} {
		var ips []netip.Addr
		for _, a := range addrs {
			ips = append(ips, netip.MustParseAddr(a))
		}
		c := NewClient(Options{Lookup: func(context.Context, string) ([]netip.Addr, error) { return ips, nil }})
		if _, err := get(t, c, "https://mixed.example/"); !errors.Is(err, ErrAddressRefused) {
			t.Errorf("%s: err = %v, want ErrAddressRefused", name, err)
		}
	}
}

// An address literal in the URL gets the same check: no lookup is
// needed to refuse https://127.0.0.1/ or https://[::1]/.
func TestAnAddressLiteralIsRefused(t *testing.T) {
	var hits atomic.Int32
	srv := counting(t, &hits, nil)
	c := NewClient(Options{TLSConfig: tlsFor(srv)})
	for _, u := range []string{
		srv.URL + "/",
		"https://192.168.1.1/",
		"https://[::1]/",
		"https://169.254.169.254/latest/meta-data/",
		"https://100.64.0.1/",
		"https://[fd00::1]/",
		"https://0.0.0.0/",
		"https://[::ffff:127.0.0.1]/",
	} {
		if _, err := get(t, c, u); !errors.Is(err, ErrAddressRefused) {
			t.Errorf("%s: err = %v, want ErrAddressRefused", u, err)
		}
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("the server was reached %d times", n)
	}
}

// The Control hook is the last word: even a dial the pre-check missed
// is refused on the address the socket is connecting to.
func TestTheDialControlRefusesARefusedAddress(t *testing.T) {
	d := &net.Dialer{Timeout: time.Second, Control: control(AddrAllowed)}
	// Lookup lies: it claims a public address was checked, and a
	// rebinding resolver then hands the dialer a loopback one. The
	// Control hook sees the real address.
	_, err := dialChecked(context.Background(), d,
		func(context.Context, string) ([]netip.Addr, error) {
			return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
		},
		func(netip.Addr) bool { return true }, // the pre-check is fooled
		"tcp", "rebind.example:443")
	if !errors.Is(err, ErrAddressRefused) {
		t.Errorf("err = %v, want the Control hook's ErrAddressRefused", err)
	}
}

func TestPlainHTTPIsRefused(t *testing.T) {
	c := NewClient(Options{Allow: exempt})
	for _, u := range []string{"http://example.com/", "http://127.0.0.1/"} {
		if _, err := get(t, c, u); !errors.Is(err, ErrNotHTTPS) {
			t.Errorf("%s: err = %v, want ErrNotHTTPS", u, err)
		}
	}
}

// The address that passed is the address dialled: one lookup, and the
// certificate (which names example.com) is checked against the URL's
// name, not the address.
func TestItDialsTheAddressItCheckedAndReturnsTheReply(t *testing.T) {
	var hits atomic.Int32
	srv := counting(t, &hits, func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	var lookups atomic.Int32
	c := NewClient(Options{
		Allow:     exempt,
		TLSConfig: tlsFor(srv),
		Lookup: func(context.Context, string) ([]netip.Addr, error) {
			if lookups.Add(1) == 1 {
				return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
			}
			return []netip.Addr{netip.MustParseAddr("10.0.0.1")}, nil
		},
	})
	port := strings.TrimPrefix(srv.URL, "https://127.0.0.1:")
	resp, err := get(t, c, "https://example.com:"+port+"/")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != "ok" {
		t.Errorf("body = %q", body)
	}
	if n := lookups.Load(); n != 1 {
		t.Errorf("%d lookups, want 1", n)
	}
}

// A redirect is not followed, wherever it points: the 3xx comes back
// to the caller and the second address is never contacted.
func TestARedirectIsNotFollowed(t *testing.T) {
	var second atomic.Int32
	target := counting(t, &second, nil)
	var first atomic.Int32
	srv := counting(t, &first, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL+"/", http.StatusFound)
	})
	c := NewClient(Options{Allow: exempt, TLSConfig: tlsFor(srv)})
	resp, err := get(t, c, srv.URL+"/")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Errorf("status = %d, want the 302 itself", resp.StatusCode)
	}
	if second.Load() != 0 {
		t.Error("the redirect target was contacted")
	}
}

func TestABodyOverTheCapFails(t *testing.T) {
	var hits atomic.Int32
	srv := counting(t, &hits, func(w http.ResponseWriter, _ *http.Request) {
		// No Content-Length: the cap must hold on a streamed body too.
		w.Header().Set("Transfer-Encoding", "chunked")
		for i := 0; i < 8; i++ {
			_, _ = w.Write([]byte(strings.Repeat("x", 512)))
			w.(http.Flusher).Flush()
		}
	})
	c := NewClient(Options{Allow: exempt, TLSConfig: tlsFor(srv), MaxResponseBytes: 1024})
	resp, err := get(t, c, srv.URL+"/")
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if _, err := io.ReadAll(resp.Body); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("read err = %v, want ErrResponseTooLarge", err)
	}

	big := counting(t, &hits, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", "4096")
		_, _ = w.Write([]byte(strings.Repeat("y", 4096)))
	})
	c = NewClient(Options{Allow: exempt, TLSConfig: tlsFor(big), MaxResponseBytes: 1024})
	if _, err := get(t, c, big.URL+"/"); !errors.Is(err, ErrResponseTooLarge) {
		t.Errorf("declared length over the cap: err = %v, want ErrResponseTooLarge", err)
	}

	exact := counting(t, &hits, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("z", 1024)))
	})
	c = NewClient(Options{Allow: exempt, TLSConfig: tlsFor(exact), MaxResponseBytes: 1024})
	resp, err = get(t, c, exact.URL+"/")
	if err != nil {
		t.Fatalf("exactly the cap: %v", err)
	}
	if b, err := io.ReadAll(resp.Body); err != nil || len(b) != 1024 {
		t.Errorf("exactly the cap: %d bytes, err %v", len(b), err)
	}
	_ = resp.Body.Close()
}

func TestASlowHostTimesOut(t *testing.T) {
	release := make(chan struct{})
	var hits atomic.Int32
	srv := counting(t, &hits, func(http.ResponseWriter, *http.Request) { <-release })
	defer close(release)
	c := NewClient(Options{Allow: exempt, TLSConfig: tlsFor(srv), Timeout: 200 * time.Millisecond})
	started := time.Now()
	if _, err := get(t, c, srv.URL+"/"); err == nil {
		t.Fatal("a host that never answers returned no error")
	}
	if took := time.Since(started); took > 5*time.Second {
		t.Errorf("took %v, want about the 200ms timeout", took)
	}
}
