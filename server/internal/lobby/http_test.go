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
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// newTestHTTPStack builds a full HTTP stack — lobby, auth, hub — on
// an httptest.Server and returns helpers for talking to it.
func newTestHTTPStack(t *testing.T) (*httptest.Server, *Lobby, auth.Authenticator) {
	srv, l, a, _ := newTestHTTPStackWithCards(t, nil)
	return srv, l, a
}

// newTestHTTPStackWithCards is like newTestHTTPStack but plugs a
// cards.Index into the lobby config so upload-deck tests can
// resolve fixture names.
func newTestHTTPStackWithCards(t *testing.T, idx *cards.Index) (*httptest.Server, *Lobby, auth.Authenticator, *cards.Index) {
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
		Cards:      idx,
	}

	mux := http.NewServeMux()
	lobbyHandler := Handler(cfg)
	mux.Handle("/", lobbyHandler)
	mux.HandleFunc("GET /ws", hub.ServeWS)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, l, a, idx
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
	_, alice, err := l.Join(m.ID, m.InviteToken, "Alice")
	if err != nil {
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

	// Satisfy Start's new "every seat has a deck" invariant.
	if _, err := l.SetDeck(m.ID, alice, "dummy", []game.Card{game.NewCommander("A", uuid.Nil), game.NewCard("F", uuid.Nil)}); err != nil {
		t.Fatalf("SetDeck Alice: %v", err)
	}
	if _, err := l.SetDeck(m.ID, joined.PlayerID, "dummy", []game.Card{game.NewCommander("B", uuid.Nil), game.NewCard("F", uuid.Nil)}); err != nil {
		t.Fatalf("SetDeck Carol: %v", err)
	}
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

func TestDeleteGameAdminOnly(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)

	// Create via admin, get token.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var adminSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&adminSess)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games", adminSess.Token, createGameRequest{Name: "FNM"})
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()

	// Join as a player so we can assert player sessions cannot DELETE.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var playerSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&playerSess)
	resp.Body.Close()

	doDelete := func(token string) *http.Response {
		req, _ := http.NewRequest(http.MethodDelete, srv.URL+"/games/"+meta.ID.String(), nil)
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, err := srv.Client().Do(req)
		if err != nil {
			t.Fatalf("delete: %v", err)
		}
		return resp
	}

	// No auth → 401.
	resp = doDelete("")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("anon delete: got %d, want 401", resp.StatusCode)
	}
	resp.Body.Close()

	// Player → 403.
	resp = doDelete(playerSess.Token)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("player delete: got %d, want 403", resp.StatusCode)
	}
	resp.Body.Close()

	// Admin → 204, game gone.
	resp = doDelete(adminSess.Token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("admin delete: got %d, want 204", resp.StatusCode)
	}
	resp.Body.Close()
	if _, err := l.Get(meta.ID); err != ErrGameNotFound {
		t.Errorf("after delete Get: got %v, want ErrGameNotFound", err)
	}

	// Second delete → 404.
	resp = doDelete(adminSess.Token)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("double delete: got %d, want 404", resp.StatusCode)
	}
	resp.Body.Close()
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

// buildMinimalDeckIndex populates a cards.Index with just enough
// cards to resolve a 100-card "all basics" Commander deck. Used by
// the upload-deck HTTP tests.
func buildMinimalDeckIndex(t *testing.T) *cards.Index {
	t.Helper()
	idx := cards.NewIndex()
	// A legal commander.
	idx.Put(cards.Card{
		ID:            uuid.New(),
		Name:          "Test Commander",
		TypeLine:      "Legendary Creature — Human Wizard",
		ColorIdentity: []string{"W"},
		Legalities:    map[string]string{"commander": "legal"},
	})
	// Plains — basic land, identity W, legal.
	idx.Put(cards.Card{
		ID:            uuid.New(),
		Name:          "Plains",
		TypeLine:      "Basic Land — Plains",
		ColorIdentity: []string{"W"},
		Legalities:    map[string]string{"commander": "legal"},
	})
	return idx
}

func TestUploadDeckHappyPath(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)

	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	// 1 commander + 99 Plains — a hand-crafted mono-white "deck".
	source := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: joined.PlayerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("upload: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var out uploadDeckResponse
	_ = json.NewDecoder(resp.Body).Decode(&out)
	resp.Body.Close()

	if out.CardCount != 100 {
		t.Errorf("card_count: got %d, want 100", out.CardCount)
	}
	if len(out.Commanders) != 1 || out.Commanders[0] != "Test Commander" {
		t.Errorf("commanders: got %v", out.Commanders)
	}
	if len(out.Game.Players) != 1 || !out.Game.Players[0].DeckUploaded {
		t.Errorf("seat deck_uploaded: got %+v", out.Game.Players)
	}
}

func TestUploadDeckPlayerCannotSetAnothersDeck(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)

	meta, _ := l.Create("FNM")
	// Seat Alice via lobby API to get her player ID directly.
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")

	// Carol joins via HTTP and gets a RolePlayer session bound to
	// her own player ID — she should be forbidden from uploading
	// Alice's deck.
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Carol"})
	var carolSession sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&carolSession)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", carolSession.Token,
		uploadDeckRequest{Format: "text", Source: "1 Test Commander\n", PlayerID: alice})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-player upload: got %d, want 403", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestUploadDeckValidationError(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)
	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	// Missing 50 cards → 422 with violations.
	source := "Commander:\n1 Test Commander\nMainboard:\n49 Plains\n"
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: joined.PlayerID})
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Errorf("undersized deck: got %d, want 422", resp.StatusCode)
	}
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	resp.Body.Close()
	if _, ok := body["violations"]; !ok {
		t.Errorf("expected violations in response body, got %+v", body)
	}
}

func TestUploadDeck503WhenIndexMissing(t *testing.T) {
	srv, l, _, _ := newTestHTTPStackWithCards(t, nil)
	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: "1 Sol Ring\n", PlayerID: joined.PlayerID})
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("no-index upload: got %d, want 503", resp.StatusCode)
	}
	resp.Body.Close()
}
