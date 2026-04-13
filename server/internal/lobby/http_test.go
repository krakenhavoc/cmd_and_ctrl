package lobby

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// newTestHTTPStack builds a full HTTP stack — lobby, auth, hub — on
// an httptest.Server and returns helpers for talking to it.
func newTestHTTPStack(t *testing.T) (*httptest.Server, *Lobby, auth.Authenticator) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})

	cfg := Config{
		Lobby:      l,
		Auth:       a,
		AdminToken: "shared-admin-token",
	}

	mux := http.NewServeMux()
	lobbyHandler := Handler(cfg)
	// The lobby handler manages its own sub-routes under /; mount it
	// at the root and layer /ws on top.
	mux.Handle("/", lobbyHandler)
	mux.HandleFunc("GET /ws", hub.ServeWS)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, l, a
}

// postJSON is a small helper for testing POSTs that take JSON bodies.
// Returns the response and decoded body (if status-ok); test cleans
// up the body itself.
func postJSON(t *testing.T, srv *httptest.Server, path, token string, body any) *http.Response {
	t.Helper()
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

func TestAdminLoginFlow(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)

	// Wrong token → 401.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "nope"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("bad token status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// Correct token → session.
	resp = postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin login status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var session sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatalf("decode: %v", err)
	}
	resp.Body.Close()
	if session.Token == "" {
		t.Error("admin login: empty token")
	}
	if session.Principal.Role != auth.RoleAdmin {
		t.Errorf("role: got %q, want %q", session.Principal.Role, auth.RoleAdmin)
	}
}

func TestCreateGameRequiresAdmin(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)

	// No credential → 401.
	resp := postJSON(t, srv, "/games", "", createGameRequest{Name: "FNM"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("no auth status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// Admin creates OK.
	resp = postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var session sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&session)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games", session.Token, createGameRequest{Name: "FNM"})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create status: got %d, want %d", resp.StatusCode, http.StatusCreated)
	}
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if meta.InviteToken == "" {
		t.Error("created game is missing invite token")
	}
}

func TestJoinFlowPublicEndpoint(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)

	// Admin creates a game.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var session sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&session)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games", session.Token, createGameRequest{Name: "FNM"})
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()

	// Unauthenticated user joins via invite.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("join status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()
	if joined.PlayerID == uuid.Nil {
		t.Error("join did not return player ID")
	}
	if joined.Principal.Role != auth.RolePlayer {
		t.Errorf("role: got %q, want %q", joined.Principal.Role, auth.RolePlayer)
	}
	if joined.Principal.GameID != meta.ID {
		t.Errorf("principal gameID: got %q, want %q", joined.Principal.GameID, meta.ID)
	}
}

func TestJoinBadInviteFails(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	m, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+m.ID.String()+"/join", "",
		joinRequest{InviteToken: "not-the-real-one", Name: "Alice"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()
}

// TestWSUpgradeWithPlayerSession drives the full onboarding flow:
// join → open WS with the session cookie → receive initial snapshot.
// This is the critical end-to-end guarantee the S04 auth seam has to
// hold — a valid session cookie is enough to open a bound WS
// connection, no separate token handling on the client.
func TestWSUpgradeWithPlayerSession(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)

	// Create a game and seat one player via the lobby API, then have
	// the second seat come in through HTTP so we exercise the join
	// endpoint end-to-end (including session issuance).
	m, _ := l.Create("FNM")
	if _, _, err := l.Join(m.ID, m.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join Alice: %v", err)
	}

	resp := postJSON(t, srv, "/games/"+m.ID.String()+"/join", "",
		joinRequest{InviteToken: m.InviteToken, Name: "Carol"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("join Carol: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	if _, err := l.Start(m.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Upgrade to WS using the session token as ?token=. Browsers use
	// cookies on same-origin WS; the query-param path covers the
	// CLI / cross-origin case.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + joined.Token
	conn, resp2, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		body := ""
		if resp2 != nil {
			b, _ := io.ReadAll(resp2.Body)
			body = string(b)
		}
		t.Fatalf("ws dial: %v (body=%s)", err, body)
	}
	defer conn.Close()

	// First frame must be a snapshot of the joined game.
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial: %v", err)
	}
	var f struct {
		V       int             `json:"v"`
		Kind    string          `json:"kind"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if f.Kind != "snapshot" {
		t.Errorf("kind: got %q, want snapshot", f.Kind)
	}
	// Payload should carry the right game ID.
	var snap struct {
		Game struct {
			ID string `json:"id"`
		} `json:"game"`
	}
	_ = json.Unmarshal(f.Payload, &snap)
	if snap.Game.ID != m.ID.String() {
		t.Errorf("game id: got %q, want %q", snap.Game.ID, m.ID)
	}
}

func TestLogoutRevokesSession(t *testing.T) {
	srv, _, a := newTestHTTPStack(t)

	// Acquire an admin session.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var s sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&s)
	resp.Body.Close()

	// Pre-check: the token is valid server-side.
	if _, err := a.Validate(nil, s.Token); err != nil {
		t.Fatalf("pre-logout validate: %v", err)
	}

	// Logout.
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/logout", nil)
	req.Header.Set("Authorization", "Bearer "+s.Token)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("logout: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("logout status: got %d, want 204", resp.StatusCode)
	}
	// Cookie should be cleared.
	var sawClear bool
	for _, c := range resp.Cookies() {
		if c.Name == auth.SessionCookie && c.MaxAge < 0 {
			sawClear = true
		}
	}
	if !sawClear {
		t.Error("logout did not emit a cookie-clearing Set-Cookie")
	}
	resp.Body.Close()

	// Post-check: the token no longer validates.
	if _, err := a.Validate(nil, s.Token); err == nil {
		t.Error("post-logout validate: token still valid")
	}

	// Logout with no credential is still 204.
	resp, err = srv.Client().Post(srv.URL+"/logout", "application/json", nil)
	if err != nil {
		t.Fatalf("anon logout: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("anon logout status: got %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestWSRejectsMissingCredential(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("dial without credential: expected error")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status: got %v, want %d", resp, http.StatusUnauthorized)
	}
}

// TestWSPlayerCannotCrossGame verifies the cross-game guard on the
// WSAuthorizer: a player session for game A cannot be reused by
// passing ?game=<B>.
func TestWSPlayerCannotCrossGame(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)

	gA, _ := l.Create("A")
	gB, _ := l.Create("B")

	resp := postJSON(t, srv, "/games/"+gA.ID.String()+"/join", "",
		joinRequest{InviteToken: gA.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + joined.Token + "&game=" + gB.ID.String()
	_, resp2, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err == nil {
		t.Fatal("cross-game dial: expected error")
	}
	if resp2 == nil || resp2.StatusCode != http.StatusForbidden {
		t.Errorf("status: got %v, want %d", resp2, http.StatusForbidden)
	}
}
