package discord

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

// Tests for ADR 0156 §1: the gate's configuration, the scope it asks
// for, the member check and the refresh-token exchange.

func envOf(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestGateFromEnvBothOrNeither(t *testing.T) {
	g, err := GateFromEnv(envOf(nil))
	if err != nil || g.Enabled() {
		t.Fatalf("neither set: %+v %v, want off and no error", g, err)
	}
	g, err = GateFromEnv(envOf(map[string]string{CommunityGuildEnv: " 111 ", PlayerRoleEnv: "222"}))
	if err != nil || !g.Enabled() || g.GuildID != "111" || g.RoleID != "222" {
		t.Fatalf("both set: %+v %v", g, err)
	}

	cases := []struct {
		name    string
		env     map[string]string
		missing string
	}{
		{"guild without role", map[string]string{CommunityGuildEnv: "111"}, PlayerRoleEnv},
		{"role without guild", map[string]string{PlayerRoleEnv: "222"}, CommunityGuildEnv},
		{"role blank", map[string]string{CommunityGuildEnv: "111", PlayerRoleEnv: "  "}, PlayerRoleEnv},
	}
	for _, tc := range cases {
		_, err := GateFromEnv(envOf(tc.env))
		if err == nil {
			t.Errorf("%s: no error, want the boot to fail", tc.name)
			continue
		}
		// The message names the variable that is missing.
		if !strings.Contains(err.Error(), tc.missing+" is not") {
			t.Errorf("%s: %v does not say %s is missing", tc.name, err, tc.missing)
		}
	}
}

func TestGateFromEnvReadsTheExemptList(t *testing.T) {
	g, err := GateFromEnv(envOf(map[string]string{
		CommunityGuildEnv: "111", PlayerRoleEnv: "222", GateExemptEnv: " 333, 444 ,,",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if !g.Exempts("333") || !g.Exempts("444") || g.Exempts("555") || g.Exempts("") {
		t.Errorf("exempt list read wrong: %+v", g)
	}
	if _, err := GateFromEnv(envOf(map[string]string{GateExemptEnv: "333,abc"})); err == nil ||
		!strings.Contains(err.Error(), GateExemptEnv) || strings.Contains(err.Error(), "abc") {
		t.Errorf("a malformed entry: err %v, want one naming %s without the value", err, GateExemptEnv)
	}
	if g, err := GateFromEnv(envOf(map[string]string{GateExemptEnv: "333"})); err != nil || g.Enabled() {
		t.Errorf("a list with the gate off: %+v %v, want off and no error", g, err)
	}
}

func TestGateFromEnvRefusesMalformedSnowflakes(t *testing.T) {
	for _, tc := range []struct {
		env  map[string]string
		name string
	}{
		{map[string]string{CommunityGuildEnv: "abc", PlayerRoleEnv: "222"}, CommunityGuildEnv},
		{map[string]string{CommunityGuildEnv: "111", PlayerRoleEnv: "2 2"}, PlayerRoleEnv},
		{map[string]string{CommunityGuildEnv: "123456789012345678901", PlayerRoleEnv: "222"}, CommunityGuildEnv},
		{map[string]string{CommunityGuildEnv: "111", PlayerRoleEnv: "-5"}, PlayerRoleEnv},
	} {
		_, err := GateFromEnv(envOf(tc.env))
		if err == nil || !strings.Contains(err.Error(), tc.name) {
			t.Errorf("%v: err %v, want one naming %s", tc.env, err, tc.name)
		}
		if err != nil && (strings.Contains(err.Error(), "abc") || strings.Contains(err.Error(), "2 2")) {
			t.Errorf("the error echoes the value: %v", err)
		}
	}
}

func TestInviteURLFromEnv(t *testing.T) {
	if u, err := InviteURLFromEnv(envOf(nil)); u != "" || err != nil {
		t.Errorf("unset: %q %v, want empty and no error", u, err)
	}
	if u, err := InviteURLFromEnv(envOf(map[string]string{InviteURLEnv: " https://discord.gg/abc123 "})); u != "https://discord.gg/abc123" || err != nil {
		t.Errorf("valid: %q %v", u, err)
	}
	for _, bad := range []string{"discord.gg/abc", "http://discord.gg/abc", "javascript:alert(1)", "https://", "https://user:pw@discord.gg/x"} {
		if _, err := InviteURLFromEnv(envOf(map[string]string{InviteURLEnv: bad})); err == nil {
			t.Errorf("%q accepted", bad)
		}
	}
}

func TestScopesFollowTheGate(t *testing.T) {
	c := Config{ClientID: "id", ClientSecret: "s", RedirectURI: "https://x/cb"}
	scope := func(c Config) string {
		u, _ := url.Parse(c.AuthorizeURL("st", "ch", PromptNone))
		return u.Query().Get("scope")
	}
	if got := scope(c); got != "identify" {
		t.Errorf("gate off: scope %q, want identify", got)
	}
	c.Gate = Gate{GuildID: "1", RoleID: "2"}
	if got := scope(c); got != "identify guilds.members.read" {
		t.Errorf("gate on: scope %q, want identify guilds.members.read", got)
	}
	if c.Scopes() != scope(c) {
		t.Error("Scopes() and the authorize URL disagree")
	}
}

// memberStub serves the member endpoint with a handler the test picks.
func memberStub(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	t.Cleanup(SwapMemberEndpointForTesting(srv.URL + "/users/@me/guilds"))
}

func gated() Config {
	return Config{ClientID: "id", ClientSecret: "s", RedirectURI: "https://x/cb", Gate: Gate{GuildID: "111", RoleID: "222"}}
}

func TestCheckMemberAnswers(t *testing.T) {
	cases := []struct {
		name   string
		status int
		body   string
		want   Access
	}{
		{"has the role", 200, `{"roles":["999","222"],"nick":"x"}`, AccessAdmitted},
		{"member without the role", 200, `{"roles":["999"]}`, AccessNoRole},
		{"no roles at all", 200, `{"roles":[]}`, AccessNoRole},
		{"not a member", 404, `{"message":"Unknown Guild","code":10004}`, AccessNotMember},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			memberStub(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/users/@me/guilds/111/member" {
					t.Errorf("path %q", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer at-1" {
					t.Errorf("authorization %q", r.Header.Get("Authorization"))
				}
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			got, err := gated().CheckMember(context.Background(), http.DefaultClient, "at-1")
			if err != nil || got != tc.want {
				t.Errorf("CheckMember = %q, %v; want %q", got, err, tc.want)
			}
		})
	}
}

func TestCheckMemberOutages(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
	}{
		{"server error", 500, ``},
		{"bad gateway", 502, ``},
		{"unauthorized", 401, `{}`},
		{"unparseable body", 200, `not json`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			memberStub(t, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			})
			_, err := gated().CheckMember(context.Background(), http.DefaultClient, "at")
			if !errors.Is(err, ErrUnavailable) {
				t.Errorf("err %v, want ErrUnavailable", err)
			}
		})
	}
}

func TestCheckMemberRetriesA429Once(t *testing.T) {
	var calls atomic.Int32
	memberStub(t, func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "0.01")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"roles":["222"]}`))
	})
	got, err := gated().CheckMember(context.Background(), http.DefaultClient, "at")
	if err != nil || got != AccessAdmitted || calls.Load() != 2 {
		t.Errorf("after one 429: %q %v, %d calls; want admitted after 2", got, err, calls.Load())
	}

	calls.Store(0)
	memberStub(t, func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	})
	_, err = gated().CheckMember(context.Background(), http.DefaultClient, "at")
	if !errors.Is(err, ErrUnavailable) || calls.Load() != 2 {
		t.Errorf("two 429s: %v after %d calls, want ErrUnavailable after 2", err, calls.Load())
	}
}

func TestCheckMemberNetworkFailureIsAnOutage(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()
	t.Cleanup(SwapMemberEndpointForTesting(base))
	if _, err := gated().CheckMember(context.Background(), http.DefaultClient, "at"); !errors.Is(err, ErrUnavailable) {
		t.Errorf("err %v, want ErrUnavailable", err)
	}
}

func refreshStub(t *testing.T, h http.HandlerFunc) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	t.Cleanup(SwapEndpointsForTesting(srv.URL+"/oauth2/token", srv.URL+"/users/@me"))
}

func TestExchangeRefreshToken(t *testing.T) {
	refreshStub(t, func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("grant_type") != "refresh_token" || r.PostForm.Get("refresh_token") != "rt-old" ||
			r.PostForm.Get("client_id") != "id" || r.PostForm.Get("client_secret") != "s" {
			t.Errorf("form %v", r.PostForm)
		}
		_, _ = w.Write([]byte(`{"access_token":"at-new","refresh_token":"rt-new","scope":"identify guilds.members.read"}`))
	})
	tr, err := gated().ExchangeRefreshToken(context.Background(), http.DefaultClient, "rt-old")
	if err != nil || tr.AccessToken != "at-new" || tr.RefreshToken != "rt-new" {
		t.Errorf("= %+v, %v", tr, err)
	}
}

func TestExchangeRefreshTokenRejectedAndOutage(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		want   error
	}{
		{400, `{"error":"invalid_grant"}`, ErrRefreshRejected},
		{401, `{"error":"invalid_client"}`, ErrRefreshRejected},
		{500, ``, ErrUnavailable},
		{503, ``, ErrUnavailable},
		{200, `{}`, ErrUnavailable},
	} {
		refreshStub(t, func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tc.status)
			_, _ = w.Write([]byte(tc.body))
		})
		_, err := gated().ExchangeRefreshToken(context.Background(), http.DefaultClient, "rt-secret")
		if !errors.Is(err, tc.want) {
			t.Errorf("status %d: err %v, want %v", tc.status, err, tc.want)
		}
		if err != nil && strings.Contains(err.Error(), "rt-secret") {
			t.Errorf("status %d: the error carries the refresh token", tc.status)
		}
	}
}
