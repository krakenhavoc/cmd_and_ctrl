package discord

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestConfigEnabled(t *testing.T) {
	cases := []struct {
		name string
		c    Config
		want bool
	}{
		{"all set", Config{ClientID: "a", ClientSecret: "b", RedirectURI: "c"}, true},
		{"missing id", Config{ClientSecret: "b", RedirectURI: "c"}, false},
		{"missing secret", Config{ClientID: "a", RedirectURI: "c"}, false},
		{"missing redirect", Config{ClientID: "a", ClientSecret: "b"}, false},
		{"empty", Config{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.c.Enabled(); got != tc.want {
				t.Errorf("Enabled(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAuthorizeURLShape(t *testing.T) {
	c := Config{ClientID: "id1", ClientSecret: "sec", RedirectURI: "https://example/cb"}
	u := c.AuthorizeURL("state-abc", "challenge-xyz")

	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	q := parsed.Query()
	if q.Get("response_type") != "code" {
		t.Errorf("response_type: %q", q.Get("response_type"))
	}
	if q.Get("client_id") != "id1" {
		t.Errorf("client_id: %q", q.Get("client_id"))
	}
	if q.Get("redirect_uri") != "https://example/cb" {
		t.Errorf("redirect_uri: %q", q.Get("redirect_uri"))
	}
	if q.Get("scope") != discordScopes {
		t.Errorf("scope: %q", q.Get("scope"))
	}
	if q.Get("state") != "state-abc" {
		t.Errorf("state: %q", q.Get("state"))
	}
	if q.Get("code_challenge") != "challenge-xyz" {
		t.Errorf("code_challenge: %q", q.Get("code_challenge"))
	}
	if q.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method: %q", q.Get("code_challenge_method"))
	}
}

func TestExchangeCodeHappyPath(t *testing.T) {
	// Stub Discord's token endpoint — echo a valid token response
	// iff the form body has the fields we expect.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		for _, k := range []string{"code", "code_verifier", "client_id", "client_secret", "redirect_uri"} {
			if r.Form.Get(k) == "" {
				t.Errorf("form field %q missing", k)
			}
		}
		if r.Form.Get("grant_type") != "authorization_code" {
			t.Errorf("grant_type: %q", r.Form.Get("grant_type"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "tok-1",
			TokenType:   "Bearer",
			ExpiresIn:   604800,
			Scope:       "identify",
		})
	}))
	defer srv.Close()

	origTokenEP := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = origTokenEP }()

	c := Config{ClientID: "id1", ClientSecret: "sec", RedirectURI: "https://example/cb"}
	tr, err := c.ExchangeCode(context.Background(), srv.Client(), "code-1", "ver-1")
	if err != nil {
		t.Fatalf("ExchangeCode: %v", err)
	}
	if tr.AccessToken != "tok-1" {
		t.Errorf("AccessToken: got %q, want tok-1", tr.AccessToken)
	}
}

func TestExchangeCodeSurfaceUpstreamError(t *testing.T) {
	// Upstream 400 with an OAuth error body — we want the error
	// + description to show up verbatim in the returned error so
	// a misconfigured redirect_uri is diagnosable without reading
	// server logs.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"Redirect URI mismatch"}`))
	}))
	defer srv.Close()

	origTokenEP := tokenEndpoint
	tokenEndpoint = srv.URL
	defer func() { tokenEndpoint = origTokenEP }()

	c := Config{ClientID: "id1", ClientSecret: "sec", RedirectURI: "https://example/cb"}
	_, err := c.ExchangeCode(context.Background(), srv.Client(), "code-1", "ver-1")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid_grant") {
		t.Errorf("error should surface OAuth error code: %v", err)
	}
	if !strings.Contains(err.Error(), "Redirect URI mismatch") {
		t.Errorf("error should surface description: %v", err)
	}
}

func TestFetchUserPrefersGlobalName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-42" {
			t.Errorf("Authorization: got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"99","username":"alice.1234","global_name":"Alice","avatar":"hashhash"}`))
	}))
	defer srv.Close()

	origUserEP := userEndpoint
	userEndpoint = srv.URL
	defer func() { userEndpoint = origUserEP }()

	c := Config{}
	u, err := c.FetchUser(context.Background(), srv.Client(), "token-42")
	if err != nil {
		t.Fatalf("FetchUser: %v", err)
	}
	if u.ID != "99" {
		t.Errorf("ID: %q", u.ID)
	}
	if u.DisplayName() != "Alice" {
		t.Errorf("DisplayName: got %q, want Alice (global_name takes priority)", u.DisplayName())
	}
}

func TestFetchUserFallsBackToUsername(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// No global_name — common for users who haven't set one.
		_, _ = w.Write([]byte(`{"id":"99","username":"alice","avatar":""}`))
	}))
	defer srv.Close()

	origUserEP := userEndpoint
	userEndpoint = srv.URL
	defer func() { userEndpoint = origUserEP }()

	u, err := Config{}.FetchUser(context.Background(), srv.Client(), "tok")
	if err != nil {
		t.Fatalf("FetchUser: %v", err)
	}
	if u.DisplayName() != "alice" {
		t.Errorf("DisplayName: got %q, want alice (fallback to username)", u.DisplayName())
	}
}

func TestFetchUserRejectsEmptyID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"username":"nameonly"}`))
	}))
	defer srv.Close()

	origUserEP := userEndpoint
	userEndpoint = srv.URL
	defer func() { userEndpoint = origUserEP }()

	_, err := Config{}.FetchUser(context.Background(), srv.Client(), "tok")
	if err == nil {
		t.Fatal("expected error for empty ID")
	}
}
