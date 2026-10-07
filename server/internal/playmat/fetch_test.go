package playmat

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"image/color"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAddrAllowed(t *testing.T) {
	refused := []string{
		"127.0.0.1", "127.8.9.10", "::1",
		"10.0.0.1", "10.255.255.255", "172.16.0.1", "172.31.255.255", "192.168.1.1",
		"169.254.169.254", "169.254.0.1", "fe80::1", "fe80::a00:27ff:fe4e:66a1",
		"100.64.0.1", "100.127.255.255", // carrier-grade NAT
		"224.0.0.1", "239.255.255.250", "ff02::1", "ff0e::1",
		"0.0.0.0", "::", "0.1.2.3",
		"fc00::1", "fd12:3456:789a::1", // IPv6 unique-local
		"255.255.255.255", "240.0.0.1",
		"192.0.2.1", "198.51.100.1", "203.0.113.1", "198.18.0.1", "2001:db8::1",
		// IPv4 spelled as IPv6.
		"::ffff:127.0.0.1", "::ffff:10.0.0.1", "::ffff:169.254.169.254",
		"64:ff9b::7f00:1", "64:ff9b::a9fe:a9fe", // NAT64 to loopback, to metadata
		"2002:7f00:1::1", "2002:a9fe:a9fe::1", // 6to4 to loopback, to metadata
		"fe80::1%eth0",
	}
	for _, s := range refused {
		ip, err := netip.ParseAddr(s)
		if err != nil {
			t.Fatalf("parse %q: %v", s, err)
		}
		if AddrAllowed(ip) {
			t.Errorf("%s is allowed, want refused", s)
		}
	}
	allowed := []string{
		"1.1.1.1", "8.8.8.8", "93.184.216.34", "172.15.0.1", "172.32.0.1", "100.63.255.255",
		"100.128.0.1", "2606:4700:4700::1111", "2001:4860:4860::8888", "::ffff:8.8.8.8",
		"64:ff9b::808:808", "2002:808:808::1",
	}
	for _, s := range allowed {
		if !AddrAllowed(netip.MustParseAddr(s)) {
			t.Errorf("%s is refused, want allowed", s)
		}
	}
	if AddrAllowed(netip.Addr{}) {
		t.Error("the zero Addr is allowed")
	}
}

func TestParseURL(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", maxURLLen)
	for _, raw := range []string{
		"", "   ", "http://example.com/a.png", "ftp://example.com/a.png", "file:///etc/passwd",
		"//example.com/a.png", "example.com/a.png", "https://", "https:///a.png",
		"https://user:pw@example.com/a.png", "javascript:alert(1)", "data:image/png;base64,AAAA", long,
	} {
		if _, err := ParseURL(raw); !errors.Is(err, ErrFetch) {
			t.Errorf("ParseURL(%q) = %v, want ErrFetch", raw, err)
		}
	}
	if u, err := ParseURL("  https://example.com/mat.png?x=1 "); err != nil || u.Host != "example.com" {
		t.Errorf("a good URL: %v, %v", u, err)
	}
}

// guard is the production policy plus one exemption: the httptest
// server's own address, so a test can reach the server it started and
// still see every other loopback address refused.
func guard(ip netip.Addr) bool {
	return ip == netip.MustParseAddr("127.0.0.1") || AddrAllowed(ip)
}

func testFetcher(srv *httptest.Server, h fetcherHooks) *Fetcher {
	if h.tlsConfig == nil {
		h.tlsConfig = srv.Client().Transport.(*http.Transport).TLSClientConfig.Clone()
		h.tlsConfig.MinVersion = tls.VersionTLS12
	}
	return newFetcher(h)
}

func goodImage(t *testing.T) []byte { return pngBytes(t, solid(8, 8, color.RGBA{1, 2, 3, 255})) }

func TestFetchReturnsTheBytes(t *testing.T) {
	want := goodImage(t)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The remote Content-Type means nothing.
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(want)
	}))
	defer srv.Close()
	f := testFetcher(srv, fetcherHooks{allow: guard})
	got, err := f.Fetch(context.Background(), srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("got %d bytes, want %d", len(got), len(want))
	}
}

func TestFetchRefusesPrivateAddressesWithTheProductionGuard(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { hits.Add(1) }))
	defer srv.Close()
	// The production policy, with only the trust root swapped.
	f := testFetcher(srv, fetcherHooks{})
	for _, target := range []string{
		srv.URL, // 127.0.0.1:port
		"https://127.0.0.1:1/",
		"https://[::1]:1/",
		"https://10.1.2.3/",
		"https://172.16.0.9/",
		"https://192.168.0.1/",
		"https://169.254.169.254/latest/meta-data/",
		"https://100.64.0.1/",
		"https://0.0.0.0/",
		"https://[fd00::1]/",
		"https://[fe80::1]/",
		"https://[::ffff:127.0.0.1]/",
		"https://2130706433/", // 127.0.0.1 as a decimal integer
		"https://0x7f.1/",
	} {
		_, err := f.Fetch(context.Background(), target)
		if !errors.Is(err, ErrFetch) {
			t.Errorf("%s: err = %v, want ErrFetch", target, err)
			continue
		}
		// The refusal says what to do, not what the guard saw.
		if strings.Contains(err.Error(), "127.0.0.1") || strings.Contains(err.Error(), "169.254") {
			t.Errorf("%s: the error leaks the address: %v", target, err)
		}
	}
	if hits.Load() != 0 {
		t.Errorf("the server was contacted %d times; the guard must refuse before connecting", hits.Load())
	}
}

// DNS rebinding: a public-looking name that resolves to a private
// address is refused where it is dialled.
func TestFetchRefusesANameThatResolvesToAPrivateAddress(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("the server must not be reached")
	}))
	defer srv.Close()
	for name, addrs := range map[string][]string{
		"loopback":  {"127.0.0.2"},
		"metadata":  {"169.254.169.254"},
		"private":   {"10.9.9.9"},
		"v6 ula":    {"fd00::5"},
		"mixed":     {"8.8.8.8", "10.0.0.1"}, // one bad address refuses the name
		"mixed rev": {"10.0.0.1", "8.8.8.8"},
	} {
		var ips []netip.Addr
		for _, a := range addrs {
			ips = append(ips, netip.MustParseAddr(a))
		}
		f := testFetcher(srv, fetcherHooks{
			allow:  guard,
			lookup: func(context.Context, string) ([]netip.Addr, error) { return ips, nil },
		})
		if _, err := f.Fetch(context.Background(), "https://rebind.example/mat.png"); !errors.Is(err, ErrFetch) {
			t.Errorf("%s: err = %v, want ErrFetch", name, err)
		}
	}
}

// The lookup is made once and the address that passed is the address
// dialled: a second lookup that answers differently changes nothing.
func TestFetchDialsTheAddressItChecked(t *testing.T) {
	want := goodImage(t)
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(want) }))
	defer srv.Close()
	var lookups atomic.Int32
	f := testFetcher(srv, fetcherHooks{
		allow: guard,
		lookup: func(context.Context, string) ([]netip.Addr, error) {
			if lookups.Add(1) == 1 {
				return []netip.Addr{netip.MustParseAddr("127.0.0.1")}, nil
			}
			return []netip.Addr{netip.MustParseAddr("10.0.0.1")}, nil
		},
	})
	port := strings.TrimPrefix(srv.URL, "https://127.0.0.1:")
	// The httptest certificate covers example.com, so the fetch
	// succeeds only by dialling the first lookup's answer.
	if _, err := f.Fetch(context.Background(), "https://example.com:"+port+"/"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if n := lookups.Load(); n != 1 {
		t.Errorf("%d lookups, want 1", n)
	}
}

func redirector(t *testing.T, to func(r *http.Request) string) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, to(r), http.StatusFound)
	}))
}

func TestFetchRefusesARedirectToAPrivateAddress(t *testing.T) {
	for _, target := range []string{
		"https://10.0.0.5/a.png",
		"https://169.254.169.254/latest/meta-data/",
		"https://[::1]:9/a.png",
		"https://127.0.0.2/a.png",
		"https://rebind.example/a.png", // a name that resolves to loopback, below
	} {
		srv := redirector(t, func(*http.Request) string { return target })
		f := testFetcher(srv, fetcherHooks{
			allow: guard,
			lookup: func(_ context.Context, host string) ([]netip.Addr, error) {
				if host == "rebind.example" {
					return []netip.Addr{netip.MustParseAddr("127.0.0.2")}, nil
				}
				return nil, errors.New("nxdomain")
			},
		})
		_, err := f.Fetch(context.Background(), srv.URL)
		srv.Close()
		if !errors.Is(err, ErrFetch) {
			t.Errorf("redirect to %s: err = %v, want ErrFetch", target, err)
		}
	}
}

func TestFetchRefusesARedirectToPlainHTTP(t *testing.T) {
	srv := redirector(t, func(*http.Request) string { return "http://example.com/a.png" })
	defer srv.Close()
	f := testFetcher(srv, fetcherHooks{allow: guard})
	if _, err := f.Fetch(context.Background(), srv.URL); !errors.Is(err, ErrFetch) {
		t.Errorf("err = %v, want ErrFetch", err)
	}
}

func TestFetchFollowsAtMostThreeRedirects(t *testing.T) {
	good := goodImage(t)
	for hops, wantOK := range map[int]bool{0: true, 1: true, 3: true, 4: false, 10: false} {
		var srv *httptest.Server
		srv = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var n int
			_, _ = fmt.Sscanf(r.URL.Path, "/hop/%d", &n)
			if n < hops {
				http.Redirect(w, r, fmt.Sprintf("/hop/%d", n+1), http.StatusFound)
				return
			}
			_, _ = w.Write(good)
		}))
		f := testFetcher(srv, fetcherHooks{allow: guard})
		_, err := f.Fetch(context.Background(), srv.URL+"/hop/0")
		srv.Close()
		if wantOK && err != nil {
			t.Errorf("%d hops: %v, want success", hops, err)
		}
		if !wantOK && !errors.Is(err, ErrFetch) {
			t.Errorf("%d hops: err = %v, want ErrFetch", hops, err)
		}
	}
}

func TestFetchSendsNoCookiesCredentialsOrReferer(t *testing.T) {
	good := goodImage(t)
	var second http.Header
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/start" {
			http.SetCookie(w, &http.Cookie{Name: "track", Value: "1", Path: "/"})
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		second = r.Header.Clone()
		_, _ = w.Write(good)
	}))
	defer srv.Close()
	f := testFetcher(srv, fetcherHooks{allow: guard})
	if _, err := f.Fetch(context.Background(), srv.URL+"/start"); err != nil {
		t.Fatal(err)
	}
	for _, h := range []string{"Cookie", "Authorization", "Referer", "Proxy-Authorization"} {
		if v := second.Get(h); v != "" {
			t.Errorf("the redirected request carried %s: %q", h, v)
		}
	}
}

func TestFetchRefusesABodyOverTheCap(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No Content-Length (chunked), so only the reader's cap can stop it.
		w.(http.Flusher).Flush()
		chunk := make([]byte, 1<<20)
		for i := 0; i < 12; i++ {
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	}))
	defer srv.Close()
	f := testFetcher(srv, fetcherHooks{allow: guard})
	if _, err := f.Fetch(context.Background(), srv.URL); !errors.Is(err, ErrTooLarge) {
		t.Errorf("err = %v, want ErrTooLarge", err)
	}
}

func TestFetchRefusesADeclaredLengthOverTheCap(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprint(MaxUploadBytes+1))
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	f := testFetcher(srv, fetcherHooks{allow: guard})
	if _, err := f.Fetch(context.Background(), srv.URL); !errors.Is(err, ErrTooLarge) {
		t.Errorf("err = %v, want ErrTooLarge", err)
	}
}

func TestFetchRefusesANon200(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	defer srv.Close()
	f := testFetcher(srv, fetcherHooks{allow: guard})
	if _, err := f.Fetch(context.Background(), srv.URL); !errors.Is(err, ErrFetch) {
		t.Errorf("err = %v, want ErrFetch", err)
	}
}

func TestFetchTimesOutOnASlowHost(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)
	f := testFetcher(srv, fetcherHooks{allow: guard})
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := f.Fetch(ctx, srv.URL); !errors.Is(err, ErrFetch) {
		t.Errorf("err = %v, want ErrFetch", err)
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Errorf("took %v", d)
	}
}
