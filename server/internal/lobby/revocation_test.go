package lobby

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/users"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// Tests for S34 sub-PR 7 at the HTTP layer (ADR 0051 decision 6, and
// decision 3's identity TTL): logout-everywhere and the admin's
// revoke-sessions withdraw every session a user holds, the old tokens
// are refused on every route and on the WS upgrade, a new sign-in
// works, and sessions with no user are never touched.

// evictCall records one EvictUserSessions call.
type evictCall struct {
	user   uuid.UUID
	before time.Time
}

type recordingEvictor struct {
	mu    sync.Mutex
	calls []evictCall
}

func (r *recordingEvictor) EvictUserSessions(user uuid.UUID, before time.Time) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, evictCall{user, before})
	return 1
}

// revocationStack is production's wiring over a real database: HMAC
// sessions wrapped with the users.Revocations the routes write to.
type revocationStack struct {
	srv     *httptest.Server
	lobby   *Lobby
	stub    *discordStub
	state   *discord.StateStore
	auth    auth.Authenticator
	evictor *recordingEvictor
}

func newRevocationStack(t *testing.T, configure func(*Config)) revocationStack {
	t.Helper()
	// These tests sign in, join and log in as admin more often than
	// the shared per-IP bucket's burst of 5 allows.
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	d := openTestDB(t, t.TempDir())
	us := users.NewSQLStore(d, nil)
	rev, err := users.NewRevocations(context.Background(), us)
	if err != nil {
		t.Fatalf("NewRevocations: %v", err)
	}
	h, err := auth.NewHMACAuthenticator([]byte("lobby-test-session-key-0123456789abcdef"))
	if err != nil {
		t.Fatalf("NewHMACAuthenticator: %v", err)
	}
	a := auth.WithRevocation(h, rev)
	ev := &recordingEvictor{}
	srv, l, stub, state := newDiscordTestStackWith(t, func(c *Config) {
		c.Lobby = NewLobbyWithStore(ws.NewRoomManager(quietLogger(), ""), NewSQLStore(d))
		c.Auth = a
		c.Users = us
		c.Revocations = rev
		c.SessionEvictor = ev
		if configure != nil {
			configure(c)
		}
	})
	return revocationStack{srv: srv, lobby: l, stub: stub, state: state, auth: a, evictor: ev}
}

// statusOf GETs /me with tok and returns the status and error text.
func statusOf(t *testing.T, srv *httptest.Server, tok string) (int, string) {
	t.Helper()
	resp := doGet(t, srv, "/me", tok)
	defer resp.Body.Close()
	var body struct {
		Error string `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	return resp.StatusCode, body.Error
}

func post(t *testing.T, srv *httptest.Server, path, tok string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func adminToken(t *testing.T, srv *httptest.Server) string {
	t.Helper()
	resp := postJSON(t, srv, "/admin/login", "", map[string]string{"token": "shared-admin-token"})
	defer resp.Body.Close()
	var body struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Token == "" {
		t.Fatalf("admin login: %d", resp.StatusCode)
	}
	return body.Token
}

func guestSeatToken(t *testing.T, srv *httptest.Server, l *Lobby) string {
	t.Helper()
	meta, err := l.Create("guest table")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	resp := postJSON(t, srv, "/join", "", map[string]string{"invite_token": meta.InviteToken, "name": "Guest"})
	defer resp.Body.Close()
	var body struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if body.Token == "" {
		t.Fatalf("guest join: %d", resp.StatusCode)
	}
	return body.Token
}

// signInAs switches the stubbed Discord account and signs in.
func signInAs(t *testing.T, s revocationStack, discordID string) string {
	t.Helper()
	s.stub.userBody = `{"id":"` + discordID + `","username":"u` + discordID + `","global_name":"U","avatar":"av"}`
	return identityTokenFromCallback(t, s.srv, s.state)
}

// afterAMillisecond waits past the revocation's millisecond. A token
// minted in the same millisecond as the revocation is refused (at or
// before), which a fast test could otherwise hit by accident.
func afterAMillisecond() { time.Sleep(2 * time.Millisecond) }

func TestLogoutEverywhereRevokesEverySessionTheUserHolds(t *testing.T) {
	s := newRevocationStack(t, nil)

	// Two browsers signed in as the same person, and a seat claimed
	// from one of them.
	laptop := signInAs(t, s, "alice")
	phone := signInAs(t, s, "alice")
	meta, _ := s.lobby.Create("FNM")
	resp := postJSON(t, s.srv, "/join", phone, map[string]string{"invite_token": meta.InviteToken})
	var joined struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()
	seat := joined.Token
	user := mustValidate(t, s.auth, laptop).UserID
	if mustValidate(t, s.auth, seat).UserID != user {
		t.Fatal("seat session is not the same user")
	}

	bob := signInAs(t, s, "bob")
	guest := guestSeatToken(t, s.srv, s.lobby)
	admin := adminToken(t, s.srv)

	resp = post(t, s.srv, "/logout/everywhere", laptop)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("POST /logout/everywhere: %d, want 204", resp.StatusCode)
	}
	cleared := false
	for _, ck := range resp.Cookies() {
		if ck.Name == auth.SessionCookie && ck.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("logout-everywhere did not clear the session cookie")
	}

	for name, tok := range map[string]string{"laptop": laptop, "phone": phone, "seat": seat} {
		if code, msg := statusOf(t, s.srv, tok); code != http.StatusUnauthorized || msg != "session revoked" {
			t.Errorf("%s token after logout-everywhere: %d %q, want 401 \"session revoked\"", name, code, msg)
		}
	}
	for name, tok := range map[string]string{"bob": bob, "guest": guest, "admin": admin} {
		if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
			t.Errorf("%s token after alice's logout-everywhere: %d, want 200", name, code)
		}
	}

	// Open sockets for that user were asked to close, with the same
	// watermark the authenticator now enforces.
	if len(s.evictor.calls) != 1 || s.evictor.calls[0].user != user {
		t.Fatalf("evictor calls = %+v, want one for %s", s.evictor.calls, user)
	}

	// Signing in again gives a working session.
	afterAMillisecond()
	fresh := signInAs(t, s, "alice")
	if code, _ := statusOf(t, s.srv, fresh); code != http.StatusOK {
		t.Errorf("new sign-in after logout-everywhere: %d, want 200", code)
	}
	if got := mustValidate(t, s.auth, fresh).UserID; got != user {
		t.Errorf("new sign-in is user %s, want the same person %s", got, user)
	}

	// The revoked token cannot sign out everywhere again: it is not a
	// session any more.
	resp = post(t, s.srv, "/logout/everywhere", laptop)
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("logout-everywhere with a revoked token: %d, want 401", resp.StatusCode)
	}
}

func TestLogoutEverywhereOnlyActsOnTheCallersOwnUser(t *testing.T) {
	s := newRevocationStack(t, nil)

	// No session.
	resp := post(t, s.srv, "/logout/everywhere", "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no session: %d, want 401", resp.StatusCode)
	}

	// Sessions with no user have nothing to revoke, and are told to
	// use /logout.
	for name, tok := range map[string]string{
		"admin": adminToken(t, s.srv),
		"guest": guestSeatToken(t, s.srv, s.lobby),
	} {
		resp := post(t, s.srv, "/logout/everywhere", tok)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s: %d, want 403", name, resp.StatusCode)
		}
		if code, _ := statusOf(t, s.srv, tok); code != http.StatusOK {
			t.Errorf("%s token after a refused logout-everywhere: %d", name, code)
		}
	}
	if len(s.evictor.calls) != 0 {
		t.Errorf("evictor called %d times for sessions with no user", len(s.evictor.calls))
	}

	// There is no way to name another user: bob signing out everywhere
	// leaves alice alone.
	alice := signInAs(t, s, "alice")
	bob := signInAs(t, s, "bob")
	resp = post(t, s.srv, "/logout/everywhere", bob)
	resp.Body.Close()
	if code, _ := statusOf(t, s.srv, alice); code != http.StatusOK {
		t.Errorf("alice after bob's logout-everywhere: %d", code)
	}
}

func TestAdminRevokeUserSessions(t *testing.T) {
	s := newRevocationStack(t, nil)
	alice := signInAs(t, s, "alice")
	aliceID := mustValidate(t, s.auth, alice).UserID
	bob := signInAs(t, s, "bob")
	admin := adminToken(t, s.srv)
	path := "/admin/users/" + aliceID.String() + "/revoke-sessions"

	// Only an admin. A user cannot revoke another user, nor use the
	// admin route on themselves.
	for name, tok := range map[string]string{"bob": bob, "alice": alice, "guest": guestSeatToken(t, s.srv, s.lobby)} {
		resp := post(t, s.srv, path, tok)
		resp.Body.Close()
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("%s calling the admin route: %d, want 403", name, resp.StatusCode)
		}
	}
	resp := post(t, s.srv, path, "")
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no session: %d, want 401", resp.StatusCode)
	}
	if code, _ := statusOf(t, s.srv, alice); code != http.StatusOK {
		t.Fatalf("alice revoked by a refused call: %d", code)
	}

	resp = post(t, s.srv, path, admin)
	var out struct {
		UserID                uuid.UUID `json:"user_id"`
		SessionsInvalidBefore time.Time `json:"sessions_invalid_before"`
		SocketsClosed         int       `json:"sockets_closed"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin revoke: %d", resp.StatusCode)
	}
	if out.UserID != aliceID || out.SessionsInvalidBefore.IsZero() || out.SocketsClosed != 1 {
		t.Errorf("response %+v", out)
	}

	if code, msg := statusOf(t, s.srv, alice); code != http.StatusUnauthorized || msg != "session revoked" {
		t.Errorf("alice after admin revoke: %d %q", code, msg)
	}
	if code, _ := statusOf(t, s.srv, bob); code != http.StatusOK {
		t.Errorf("bob after alice's revoke: %d", code)
	}
	// The admin's own session is not a user session and is untouched.
	if code, _ := statusOf(t, s.srv, admin); code != http.StatusOK {
		t.Errorf("admin after revoking: %d", code)
	}

	// Revocation only: alice can sign in again.
	afterAMillisecond()
	if code, _ := statusOf(t, s.srv, signInAs(t, s, "alice")); code != http.StatusOK {
		t.Errorf("alice signing in again after removal: %d", code)
	}

	for p, want := range map[string]int{
		"/admin/users/" + uuid.New().String() + "/revoke-sessions": http.StatusNotFound,
		"/admin/users/not-a-uuid/revoke-sessions":                  http.StatusBadRequest,
		"/admin/users/" + uuid.Nil.String() + "/revoke-sessions":   http.StatusBadRequest,
	} {
		resp := post(t, s.srv, p, admin)
		resp.Body.Close()
		if resp.StatusCode != want {
			t.Errorf("POST %s: %d, want %d", p, resp.StatusCode, want)
		}
	}
}

// TestRevokedSessionCannotOpenAWebSocket is decision 6 on the upgrade:
// the WS authorizer validates through the same wrapped authenticator,
// and a binding carries the session's user and issue time so the hub
// can evict it.
func TestRevokedSessionCannotOpenAWebSocket(t *testing.T) {
	s := newRevocationStack(t, nil)
	meta, _ := s.lobby.Create("FNM")
	idTok := signInAs(t, s, "alice")
	resp := postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken})
	var joined struct {
		Token string `json:"token"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	authz := &WSAuthorizer{Auth: s.auth}
	upgrade := func() (ws.Binding, error) {
		r := httptest.NewRequest(http.MethodGet, "/ws?game="+meta.ID.String(), nil)
		r.Header.Set("Authorization", "Bearer "+joined.Token)
		return authz.AuthorizeUpgrade(r)
	}
	b, err := upgrade()
	if err != nil {
		t.Fatalf("upgrade before revoke: %v", err)
	}
	seat := mustValidate(t, s.auth, joined.Token)
	if b.UserID != seat.UserID || !b.IssuedAt.Equal(seat.IssuedAt) {
		t.Errorf("binding carries user %s at %v, want %s at %v", b.UserID, b.IssuedAt, seat.UserID, seat.IssuedAt)
	}

	resp = post(t, s.srv, "/logout/everywhere", idTok)
	resp.Body.Close()
	if _, err := upgrade(); err == nil {
		t.Fatal("revoked seat session opened a WebSocket")
	}
	if _, err := s.auth.Validate(context.Background(), joined.Token); !errors.Is(err, auth.ErrRevokedCredential) {
		t.Errorf("Validate: %v, want ErrRevokedCredential", err)
	}
}

// TestSessionTTLSelection is decision 3: the identity-only session a
// Discord sign-in mints lives IdentityTTL (30 days by default); every
// session minted from it, and every other session, lives SessionTTL.
func TestSessionTTLSelection(t *testing.T) {
	lifetime := func(t *testing.T, a auth.Authenticator, tok string) time.Duration {
		t.Helper()
		p := mustValidate(t, a, tok)
		return p.ExpiresAt.Sub(p.IssuedAt)
	}

	t.Run("defaults", func(t *testing.T) {
		s := newRevocationStack(t, nil)
		idTok := signInAs(t, s, "alice")
		if got := lifetime(t, s.auth, idTok); got != 30*24*time.Hour {
			t.Errorf("identity session lives %v, want 720h", got)
		}
		meta, _ := s.lobby.Create("FNM")
		resp := postJSON(t, s.srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken})
		var joined struct {
			Token string `json:"token"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&joined)
		resp.Body.Close()
		if got := lifetime(t, s.auth, joined.Token); got != 12*time.Hour {
			t.Errorf("seat session from /join lives %v, want the 12h session TTL", got)
		}
		if got := lifetime(t, s.auth, adminToken(t, s.srv)); got != 12*time.Hour {
			t.Errorf("admin session lives %v, want 12h", got)
		}
	})

	t.Run("configured", func(t *testing.T) {
		s := newRevocationStack(t, func(c *Config) {
			c.SessionTTL = 2 * time.Hour
			c.IdentityTTL = 7 * 24 * time.Hour
		})
		if got := lifetime(t, s.auth, signInAs(t, s, "alice")); got != 7*24*time.Hour {
			t.Errorf("identity session lives %v, want 168h", got)
		}

		// The invite-link flow claims a seat in the callback itself:
		// that is a seat session, so SessionTTL.
		meta, _ := s.lobby.Create("FNM")
		st, _, err := s.state.Start(meta.ID, meta.InviteToken)
		if err != nil {
			t.Fatalf("Start: %v", err)
		}
		frag := fragmentParams(t, followOneRedirect(t, s.srv, "/auth/discord/callback?state="+st+"&code=code-3"))
		if got := lifetime(t, s.auth, frag.Get("token")); got != 2*time.Hour {
			t.Errorf("invite-flow seat session lives %v, want 2h", got)
		}
		if got := lifetime(t, s.auth, guestSeatToken(t, s.srv, s.lobby)); got != 2*time.Hour {
			t.Errorf("guest seat session lives %v, want 2h", got)
		}
	})
}

// TestOAuthFragmentCarriesTheUserID: the client builds its principal
// from the fragment, and user_id is what tells it "sign out everywhere"
// will work.
func TestOAuthFragmentCarriesTheUserID(t *testing.T) {
	s := newRevocationStack(t, nil)
	st, _, _ := s.state.Start(uuid.Nil, "")
	frag := fragmentParams(t, followOneRedirect(t, s.srv, "/auth/discord/callback?state="+st+"&code=code-1"))
	p := mustValidate(t, s.auth, frag.Get("token"))
	if frag.Get("user_id") != p.UserID.String() {
		t.Errorf("fragment user_id %q, want %s", frag.Get("user_id"), p.UserID)
	}

	// No database, no user, no user_id.
	srv, _, _, state := newDiscordTestStack(t)
	st, _, _ = state.Start(uuid.Nil, "")
	frag = fragmentParams(t, followOneRedirect(t, srv, "/auth/discord/callback?state="+st+"&code=code-1"))
	if frag.Has("user_id") {
		t.Errorf("fragment carries user_id %q with no database", frag.Get("user_id"))
	}
}

// TestRevocationWithoutADatabase is CMDCTRL_DATA_DIR="": no revocation
// list, so the routes say so, and everything else works as before.
func TestRevocationWithoutADatabase(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, l, _, state := newDiscordTestStack(t)
	idTok := identityTokenFromCallback(t, srv, state)
	admin := adminToken(t, srv)

	// The session has no user: nothing to sign out of everywhere.
	resp := post(t, srv, "/logout/everywhere", idTok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("logout-everywhere with no database: %d, want 403", resp.StatusCode)
	}
	resp = post(t, srv, "/admin/users/"+uuid.New().String()+"/revoke-sessions", admin)
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("admin revoke with no database: %d, want 503", resp.StatusCode)
	}

	// Everything else is unchanged: the identity session validates,
	// joins a table, and plain /logout works.
	if code, _ := statusOf(t, srv, idTok); code != http.StatusOK {
		t.Errorf("/me: %d", code)
	}
	meta, _ := l.Create("FNM")
	resp = postJSON(t, srv, "/join", idTok, map[string]string{"invite_token": meta.InviteToken})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/join: %d", resp.StatusCode)
	}
	resp = post(t, srv, "/logout", idTok)
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("/logout: %d", resp.StatusCode)
	}
}

// A database-backed user session on a server wired without a revoker
// (never production, which wires both or neither) gets the 503, not a
// silent success.
func TestLogoutEverywhereWithoutARevokerIs503(t *testing.T) {
	s := newRevocationStack(t, func(c *Config) { c.Revocations = nil })
	resp := post(t, s.srv, "/logout/everywhere", signInAs(t, s, "alice"))
	resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("status %d, want 503", resp.StatusCode)
	}
}
