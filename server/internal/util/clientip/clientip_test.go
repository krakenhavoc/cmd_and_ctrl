package clientip

import (
	"net/http/httptest"
	"testing"
)

// TestHeaderChain pins what the server reads from the chain the hosts
// actually run (#3334): Cloudflare edge -> cloudflared -> Caddy, with
// Caddy trusting its loopback peer. The Caddy rows are what Caddy
// 2.11 was observed to forward in a local reproduction: untrusted, it
// replaces X-Forwarded-For with its peer; trusted, it appends the peer.
// CF-Connecting-IP passes through either way.
func TestHeaderChain(t *testing.T) {
	cases := []struct {
		name   string
		trust  bool
		remote string
		cf     string
		xff    []string
		want   string
	}{
		{
			name:   "today's hosts: Caddy untrusting, CF-Connecting-IP wins",
			trust:  true,
			remote: "127.0.0.1:51234",
			cf:     "203.0.113.7",
			xff:    []string{"127.0.0.1"},
			want:   "203.0.113.7",
		},
		{
			name:   "Caddy trusting loopback, spoofed leftmost entry",
			trust:  true,
			remote: "127.0.0.1:51234",
			cf:     "203.0.113.7",
			xff:    []string{"6.6.6.6, 203.0.113.7, 127.0.0.1"},
			want:   "203.0.113.7",
		},
		{
			name:   "no CF-Connecting-IP: rightmost non-loopback XFF entry",
			trust:  true,
			remote: "127.0.0.1:51234",
			xff:    []string{"6.6.6.6, 203.0.113.7, 127.0.0.1"},
			want:   "203.0.113.7",
		},
		{
			name:   "no CF-Connecting-IP, cloudflared reached Caddy over ::1",
			trust:  true,
			remote: "127.0.0.1:51234",
			xff:    []string{"6.6.6.6, 203.0.113.7, ::1"},
			want:   "203.0.113.7",
		},
		{
			name:   "XFF split over two header lines",
			trust:  true,
			remote: "127.0.0.1:51234",
			xff:    []string{"6.6.6.6", "203.0.113.7, 127.0.0.1"},
			want:   "203.0.113.7",
		},
		{
			name:   "single non-loopback proxy appends the client",
			trust:  true,
			remote: "10.0.0.1:4242",
			xff:    []string{"spoofed.example, 198.51.100.9"},
			want:   "198.51.100.9",
		},
		{
			name:   "garbage right of the client stops the walk",
			trust:  true,
			remote: "10.0.0.1:4242",
			xff:    []string{"198.51.100.9, not-an-ip"},
			want:   "10.0.0.1",
		},
		{
			name:   "only loopback in XFF: RemoteAddr",
			trust:  true,
			remote: "127.0.0.1:51234",
			xff:    []string{"127.0.0.1"},
			want:   "127.0.0.1",
		},
		{
			name:   "unparseable CF-Connecting-IP is ignored",
			trust:  true,
			remote: "127.0.0.1:51234",
			cf:     "unknown",
			xff:    []string{"203.0.113.7, 127.0.0.1"},
			want:   "203.0.113.7",
		},
		{
			name:   "trusted, no headers: RemoteAddr host",
			trust:  true,
			remote: "10.0.0.2:9999",
			want:   "10.0.0.2",
		},
		{
			name:   "flag off: headers ignored",
			trust:  false,
			remote: "10.0.0.1:4242",
			cf:     "203.0.113.7",
			xff:    []string{"203.0.113.8"},
			want:   "10.0.0.1",
		},
		{
			name:   "IPv6 CF-Connecting-IP",
			trust:  true,
			remote: "127.0.0.1:51234",
			cf:     "2001:db8::7",
			want:   "2001:db8::7",
		},
		{
			name:   "IPv6 written in full folds to its short form",
			trust:  true,
			remote: "127.0.0.1:51234",
			cf:     "2001:0db8:0000:0000:0000:0000:0000:0007",
			want:   "2001:db8::7",
		},
		{
			name:   "IPv4-mapped IPv6 folds to IPv4",
			trust:  true,
			remote: "127.0.0.1:51234",
			cf:     "::ffff:203.0.113.7",
			want:   "203.0.113.7",
		},
		{
			name:   "bracketed IPv6 with a port in XFF",
			trust:  true,
			remote: "127.0.0.1:51234",
			xff:    []string{"[2001:db8::7]:443, ::1"},
			want:   "2001:db8::7",
		},
		{
			name:   "IPv4 with a port in XFF",
			trust:  true,
			remote: "127.0.0.1:51234",
			xff:    []string{" 203.0.113.7:8443 ,127.0.0.1"},
			want:   "203.0.113.7",
		},
		{
			name:   "IPv6 RemoteAddr, port stripped",
			trust:  false,
			remote: "[2001:db8::9]:5555",
			want:   "2001:db8::9",
		},
		{
			name:   "IPv6 RemoteAddr with a zone",
			trust:  false,
			remote: "[fe80::1%eth0]:5555",
			want:   "fe80::1",
		},
		{
			name:   "RemoteAddr that is not an address is kept",
			trust:  false,
			remote: "pipe",
			want:   "pipe",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tc.remote
			if tc.cf != "" {
				r.Header.Set("CF-Connecting-IP", tc.cf)
			}
			for _, v := range tc.xff {
				r.Header.Add("X-Forwarded-For", v)
			}
			if got := resolve(r, tc.trust); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

// TestFromRequestReadsTheFlag checks the exported entry point follows
// CMDCTRL_TRUST_FORWARDED, read per call.
func TestFromRequestReadsTheFlag(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	r.RemoteAddr = "127.0.0.1:51234"
	r.Header.Set("CF-Connecting-IP", "203.0.113.7")

	for _, v := range []string{"", "0", "false"} {
		t.Setenv(EnvTrustForwarded, v)
		if got := FromRequest(r); got != "127.0.0.1" {
			t.Errorf("%s=%q: got %q, want 127.0.0.1", EnvTrustForwarded, v, got)
		}
	}
	t.Setenv(EnvTrustForwarded, "1")
	if got := FromRequest(r); got != "203.0.113.7" {
		t.Errorf("%s=1: got %q, want 203.0.113.7", EnvTrustForwarded, got)
	}
}
