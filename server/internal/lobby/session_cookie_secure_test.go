package lobby

import (
	"net/http/httptest"
	"testing"
	"time"
)

// TestSessionCookieSecureFlag pins both session cookies to the
// secureCookies decision (#4204).
func TestSessionCookieSecureFlag(t *testing.T) {
	for _, tc := range []struct {
		uri  string
		want bool
	}{{"https://c2.example/cb", true}, {"", false}} {
		cfg := Config{}
		cfg.Discord.RedirectURI = tc.uri
		t.Setenv("CMDCTRL_SECURE_COOKIES", "")
		w := httptest.NewRecorder()
		setSessionCookie(cfg, w, "tok", time.Now().Add(time.Hour))
		clearSessionCookie(cfg, w)
		cs := w.Result().Cookies()
		if len(cs) != 2 {
			t.Fatalf("cookies = %d, want 2", len(cs))
		}
		for _, c := range cs {
			if c.Secure != tc.want {
				t.Errorf("uri %q: Secure = %v, want %v", tc.uri, c.Secure, tc.want)
			}
		}
	}
}
