package lobby

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// newDiscordTestStack stands up a lobby HTTP server with a stubbed
// Discord backend so tests can drive /auth/discord/callback without
// touching the real Discord API. The returned discordSrv is the
// stub Discord; tests can swap behaviours by overriding its
// handler before calling /callback.
type discordStub struct {
	// mu guards the fields the handlers read and write, for tests that
	// change the stub between requests (setUser, tokenWasCalled).
	mu          sync.Mutex
	srv         *httptest.Server
	tokenCalled bool
	userCalled  bool
	tokenStatus int
	tokenBody   string
	userStatus  int
	userBody    string
}

// setUser changes the Discord account /users/@me reports.
func (s *discordStub) setUser(body string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.userBody = body
}

// tokenWasCalled reports whether the token endpoint was hit since the
// last call, and resets the flag.
func (s *discordStub) tokenWasCalled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	called := s.tokenCalled
	s.tokenCalled = false
	return called
}

func newDiscordTestStack(t *testing.T) (*httptest.Server, *Lobby, *discordStub, *discord.StateStore) {
	t.Helper()
	return newDiscordTestStackWith(t, nil)
}

// newDiscordTestStackWith is newDiscordTestStack with a hook that may
// rewrite the Config — swap the Lobby, the Authenticator, or wire a
// user store — before the handler is built. The returned *Lobby is
// whatever cfg.Lobby ends up as.
func newDiscordTestStackWith(t *testing.T, configure func(*Config)) (*httptest.Server, *Lobby, *discordStub, *discord.StateStore) {
	t.Helper()
	stub := &discordStub{
		tokenStatus: http.StatusOK,
		tokenBody:   `{"access_token":"tok-stub","token_type":"Bearer","expires_in":3600,"scope":"identify"}`,
		userStatus:  http.StatusOK,
		userBody:    `{"id":"discord-99","username":"alice","global_name":"Alice","avatar":"avhash"}`,
	}
	stubMux := http.NewServeMux()
	stubMux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, _ *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.tokenCalled = true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.tokenStatus)
		_, _ = w.Write([]byte(stub.tokenBody))
	})
	stubMux.HandleFunc("/users/@me", func(w http.ResponseWriter, _ *http.Request) {
		stub.mu.Lock()
		defer stub.mu.Unlock()
		stub.userCalled = true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(stub.userStatus)
		_, _ = w.Write([]byte(stub.userBody))
	})
	stub.srv = httptest.NewServer(stubMux)
	t.Cleanup(stub.srv.Close)

	// Point the discord package at the stub.
	t.Cleanup(swapDiscordEndpoints(stub.srv.URL+"/oauth2/token", stub.srv.URL+"/users/@me"))

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()

	store := discord.NewStateStore()
	cfg := Config{
		Lobby:             l,
		Auth:              a,
		AdminToken:        "shared-admin-token",
		Discord:           discord.Config{ClientID: "id", ClientSecret: "sec", RedirectURI: "http://localhost/cb"},
		DiscordStateStore: store,
		DiscordHTTPClient: stub.srv.Client(),
	}
	if configure != nil {
		configure(&cfg)
	}
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, cfg.Lobby, stub, store
}

func TestDiscordConfigReportsEnabled(t *testing.T) {
	srv, _, _, _ := newDiscordTestStack(t)

	resp := doGet(t, srv, "/auth/discord/config", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d", resp.StatusCode)
	}
	var body map[string]bool
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if !body["enabled"] {
		t.Errorf("enabled: got false, want true")
	}
}

func TestDiscordConfigReportsDisabledWhenUnset(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)
	resp := doGet(t, srv, "/auth/discord/config", "")
	defer resp.Body.Close()
	var body map[string]bool
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body["enabled"] {
		t.Errorf("enabled: got true, want false (no Discord config wired in default test stack)")
	}
}

func TestDiscordStartRedirectsToDiscord(t *testing.T) {
	srv, l, _, _ := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")

	// http.Client follows redirects by default; we want to see the
	// 302 itself, so install a CheckRedirect that aborts the chain.
	client := *srv.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Get(srv.URL + "/auth/discord/start?game=" + meta.ID.String() + "&t=" + meta.InviteToken)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status: got %d, want 302", resp.StatusCode)
	}
	loc := resp.Header.Get("Location")
	if !strings.Contains(loc, "discord.com/oauth2/authorize") {
		t.Errorf("Location: %q does not point to Discord authorize", loc)
	}
	if !strings.Contains(loc, "code_challenge_method=S256") {
		t.Errorf("Location must request PKCE S256: %q", loc)
	}
}

func TestDiscordStartReturns503WhenDisabled(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")
	resp := doGet(t, srv, "/auth/discord/start?game="+meta.ID.String()+"&t="+meta.InviteToken, "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status: got %d, want 503", resp.StatusCode)
	}
}

func TestDiscordCallbackHappyPath(t *testing.T) {
	srv, l, stub, store := newDiscordTestStack(t)
	meta, _ := l.Create("FNM")

	// Park a state ourselves so we can drive /callback directly
	// without re-walking /start.
	state, _, err := store.Start(meta.ID, meta.InviteToken)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	client := *srv.Client()
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Get(srv.URL + "/auth/discord/callback?state=" + state + "&code=code-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("status: got %d, want 302", resp.StatusCode)
	}
	if !stub.tokenCalled {
		t.Error("token endpoint not called")
	}
	if !stub.userCalled {
		t.Error("/users/@me not called")
	}
	loc := resp.Header.Get("Location")
	if !strings.HasPrefix(loc, "/#/oauth-complete?") {
		t.Fatalf("Location should redirect to SPA oauth-complete route: %q", loc)
	}
	for _, want := range []string{"token=", "game=" + meta.ID.String(), "player_id=", "expires_at="} {
		if !strings.Contains(loc, want) {
			t.Errorf("Location missing %q: %q", want, loc)
		}
	}

	// Confirm the seat was actually claimed with the Discord
	// identity bound — opponents should see Alice's display name
	// + avatar hash on the lobby SeatInfo.
	updated, err := l.Get(meta.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(updated.Players) != 1 {
		t.Fatalf("seat count: got %d", len(updated.Players))
	}
	p := updated.Players[0]
	if p.DisplayName != "Alice" {
		t.Errorf("DisplayName: got %q, want Alice", p.DisplayName)
	}
	if p.DiscordID != "discord-99" {
		t.Errorf("DiscordID: got %q", p.DiscordID)
	}
	if p.DiscordAvatarHash != "avhash" {
		t.Errorf("DiscordAvatarHash: got %q", p.DiscordAvatarHash)
	}
}

func TestDiscordCallbackRejectsForgedState(t *testing.T) {
	srv, _, _, _ := newDiscordTestStack(t)
	resp := doGet(t, srv, "/auth/discord/callback?state=fake&code=code-1", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", resp.StatusCode)
	}
}

func TestDiscordCallbackSurfacesUserDeny(t *testing.T) {
	srv, _, _, _ := newDiscordTestStack(t)
	// Discord bounces back with ?error=access_denied when the user
	// clicks Cancel on the consent screen. We want a clear 400 with
	// the reason rather than a confusing 500.
	resp := doGet(t, srv, "/auth/discord/callback?error=access_denied&error_description=The+user+denied+the+request", "")
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status: got %d, want 400", resp.StatusCode)
	}
}

// swapDiscordEndpoints points the discord package's
// authorize/token/user URLs at a stub for the duration of the
// test. Returns a teardown function for t.Cleanup.
func swapDiscordEndpoints(token, user string) func() {
	return discord.SwapEndpointsForTesting(token, user)
}
