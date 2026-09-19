package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
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
	a := auth.NewMemoryAuthenticator()
	srv, l := newTestHTTPStackWithAuth(t, idx, a)
	return srv, l, a, idx
}

// newTestHTTPStackWithAuth is newTestHTTPStackWithCards with the
// authenticator chosen by the caller — the HMAC-session tests run the
// real handlers against auth.HMACAuthenticator.
func newTestHTTPStackWithAuth(t *testing.T, idx *cards.Index, a auth.Authenticator) (*httptest.Server, *Lobby) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
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
	return srv, l
}

// postJSON is a small helper for testing POSTs that take JSON bodies.
// Returns the response and decoded body (if status-ok); test cleans
// up the body itself.
func doGet(t *testing.T, srv *httptest.Server, path, token string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, srv.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	return resp
}

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

func TestSpectateFlowMintsSpectatorSession(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)

	// Admin creates a game; the meta carries both invites.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var session sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&session)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games", session.Token, createGameRequest{Name: "FNM"})
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()
	if meta.SpectatorInvite == "" {
		t.Fatal("Create did not mint a spectator invite")
	}
	if meta.SpectatorInvite == meta.InviteToken {
		t.Fatal("spectator invite must differ from player invite")
	}

	// Anyone with the spectator invite can spectate without claiming a
	// seat. Returns a RoleSpectator session bound to the game (no
	// player ID).
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: "watcher"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("spectate status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var spec sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&spec)
	resp.Body.Close()
	if spec.Principal.Role != auth.RoleSpectator {
		t.Errorf("role: got %q, want %q", spec.Principal.Role, auth.RoleSpectator)
	}
	if spec.Principal.GameID != meta.ID {
		t.Errorf("principal gameID: got %q, want %q", spec.Principal.GameID, meta.ID)
	}
	if spec.PlayerID != uuid.Nil {
		t.Errorf("spectator session leaked a player id: %v", spec.PlayerID)
	}
	// Returned meta should NOT carry either invite — spectators
	// shouldn't be able to forward a seat-claim invite to someone else.
	if spec.Game.InviteToken != "" || spec.Game.SpectatorInvite != "" {
		t.Errorf("spectator session response leaked invite tokens: player=%q spectator=%q",
			spec.Game.InviteToken, spec.Game.SpectatorInvite)
	}

	// Wrong invite is rejected.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: "wrong"})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("wrong invite status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// The player invite is NOT a spectator invite — distinct tokens.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.InviteToken})
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("player invite for spectate status: got %d, want %d",
			resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()
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

// TestPreviewShowsTheTableToInviteHolders covers GET /games/{id}/preview:
// the player invite and the spectator invite both unlock a scrubbed
// view of the table (no tokens, no player IDs, no Discord identity),
// a wrong or missing invite is 401, and the kind reports which
// invite was used.
func TestPreviewShowsTheTableToInviteHolders(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	m, _ := l.Create("FNM")
	if _, _, err := l.Join(m.ID, m.InviteToken, "Alice"); err != nil {
		t.Fatalf("Join Alice: %v", err)
	}
	base := "/games/" + m.ID.String() + "/preview"

	get := func(q string) (*http.Response, previewResponse) {
		resp, err := http.Get(srv.URL + base + q)
		if err != nil {
			t.Fatalf("GET preview: %v", err)
		}
		var body previewResponse
		if resp.StatusCode == http.StatusOK {
			_ = json.NewDecoder(resp.Body).Decode(&body)
		}
		resp.Body.Close()
		return resp, body
	}

	resp, body := get("?t=" + m.InviteToken)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("player invite: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if body.Invite != PreviewPlayer {
		t.Errorf("invite kind: got %q, want %q", body.Invite, PreviewPlayer)
	}
	if body.MaxSeats != game.MaxPlayers {
		t.Errorf("max_seats: got %d, want %d", body.MaxSeats, game.MaxPlayers)
	}
	if body.Game.Name != "FNM" || body.Game.State != "lobby" {
		t.Errorf("meta: got name=%q state=%q", body.Game.Name, body.Game.State)
	}
	if body.Game.InviteToken != "" || body.Game.SpectatorInvite != "" {
		t.Error("preview leaked an invite token")
	}
	if len(body.Game.Players) != 1 || body.Game.Players[0].Name != "Alice" {
		t.Fatalf("players: got %+v, want Alice", body.Game.Players)
	}
	if body.Game.Players[0].PlayerID != uuid.Nil {
		t.Error("preview leaked a player ID")
	}

	resp, body = get("?t=" + m.SpectatorInvite)
	if resp.StatusCode != http.StatusOK || body.Invite != PreviewSpectator {
		t.Errorf("spectator invite: got %d / %q, want 200 / %q", resp.StatusCode, body.Invite, PreviewSpectator)
	}

	for _, q := range []string{"", "?t=", "?t=not-the-real-one"} {
		resp, _ = get(q)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("preview %q: got %d, want %d", q, resp.StatusCode, http.StatusUnauthorized)
		}
	}
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

// TestSpectatorWSCanWatchButNotMutate end-to-ends the spectator path:
// a viewer hits POST /games/{id}/spectate, opens the WS with the
// returned session, gets a snapshot, and is rejected when they try
// to send an action frame. Confirms ReadOnly plumbing from auth →
// ws.Binding → Client.readOnly → handleAction reaches all the way
// through.
func TestSpectatorWSCanWatchButNotMutate(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)

	m, _ := l.Create("FNM")
	_, alice, err := l.Join(m.ID, m.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("Join Alice: %v", err)
	}
	_, bob, err := l.Join(m.ID, m.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("Join Bob: %v", err)
	}
	if _, err := l.SetDeck(m.ID, alice, "dummy", []game.Card{game.NewCommander("A", uuid.Nil), game.NewCard("F", uuid.Nil)}); err != nil {
		t.Fatalf("SetDeck Alice: %v", err)
	}
	if _, err := l.SetDeck(m.ID, bob, "dummy", []game.Card{game.NewCommander("B", uuid.Nil), game.NewCard("F", uuid.Nil)}); err != nil {
		t.Fatalf("SetDeck Bob: %v", err)
	}
	if _, err := l.Start(m.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Spectate via HTTP.
	resp := postJSON(t, srv, "/games/"+m.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: m.SpectatorInvite, Name: "watcher"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("spectate status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var spec sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&spec)
	resp.Body.Close()

	// Open the WS as the spectator.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + spec.Token
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	// First frame is the initial snapshot — spectator can read.
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}
	var f struct {
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal(raw, &f)
	if f.Kind != "snapshot" {
		t.Errorf("first frame kind: got %q, want snapshot", f.Kind)
	}

	// Send any action frame — must be rejected with bad_request.
	actionFrame := []byte(`{"v":0,"kind":"action","id":"00000000-0000-4000-8000-000000000001","payload":{"type":"draw_card"}}`)
	if err := conn.WriteMessage(websocket.TextMessage, actionFrame); err != nil {
		t.Fatalf("write action: %v", err)
	}
	_, raw, err = conn.ReadMessage()
	if err != nil {
		t.Fatalf("read action response: %v", err)
	}
	var resp2 struct {
		Kind    string `json:"kind"`
		Payload struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"payload"`
	}
	_ = json.Unmarshal(raw, &resp2)
	if resp2.Kind != "error" {
		t.Errorf("expected error frame for spectator action, got kind=%q", resp2.Kind)
	}
	if resp2.Payload.Code != "bad_request" {
		t.Errorf("error code: got %q, want bad_request", resp2.Payload.Code)
	}
}

// TestSpectatorBeforePlayersDoesNotCorruptState is a direct repro for
// the manual-test report: "if a spectator joins before the players,
// the players join as spectators and the game is marked as started
// prematurely". This test asserts that the server-side flow is
// independent: a spectator joining first doesn't touch game state,
// doesn't change the response of subsequent player joins, and the
// game stays in lobby until Start fires.
//
// If this test passes, the bug is client-side (likely a single-
// browser-profile localStorage session collision). If it fails, the
// server has a leak we need to chase.
func TestSpectatorBeforePlayersDoesNotCorruptState(t *testing.T) {
	srv, _, _ := newTestHTTPStack(t)

	// Admin creates a game.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var adminSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&adminSess)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games", adminSess.Token, createGameRequest{Name: "FNM"})
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()

	// Spectator joins first.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: meta.SpectatorInvite, Name: "watcher"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("spectate status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var spec sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&spec)
	resp.Body.Close()
	if spec.Principal.Role != auth.RoleSpectator {
		t.Fatalf("spectator role: got %q, want %q", spec.Principal.Role, auth.RoleSpectator)
	}

	// Game state must still be lobby — Spectate is read-only.
	resp = doGet(t, srv, "/games/"+meta.ID.String(), adminSess.Token)
	var afterSpec GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&afterSpec)
	resp.Body.Close()
	if afterSpec.State != "lobby" {
		t.Fatalf("state after spectator joined: got %q, want %q", afterSpec.State, "lobby")
	}
	if len(afterSpec.Players) != 0 {
		t.Errorf("spectator should not have created a seat; got %d players", len(afterSpec.Players))
	}

	// Now a player joins via the regular invite. They MUST get a
	// RolePlayer session, not a spectator one.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("join after spectator status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var alice sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&alice)
	resp.Body.Close()
	if alice.Principal.Role != auth.RolePlayer {
		t.Errorf("player role after spectator joined first: got %q, want %q",
			alice.Principal.Role, auth.RolePlayer)
	}
	if alice.PlayerID == uuid.Nil {
		t.Error("player session has no PlayerID")
	}

	// Game state must still be lobby (haven't called Start).
	resp = doGet(t, srv, "/games/"+meta.ID.String(), adminSess.Token)
	var afterPlayer GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&afterPlayer)
	resp.Body.Close()
	if afterPlayer.State != "lobby" {
		t.Errorf("state after player joined: got %q, want %q", afterPlayer.State, "lobby")
	}
	if len(afterPlayer.Players) != 1 {
		t.Errorf("seat count after player joined: got %d, want 1", len(afterPlayer.Players))
	}
}

// TestReplayDownloadWithDumpDir stands up a real lobby with a dump
// directory and confirms the download endpoint's behavior.
func TestReplayDownloadWithDumpDir(t *testing.T) {
	tmpDir := t.TempDir()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, tmpDir)
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})
	cfg := Config{Lobby: l, Auth: a, AdminToken: "shared-admin-token"}

	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Admin creates a game, gets an admin session token.
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	var adminSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&adminSess)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games", adminSess.Token, createGameRequest{Name: "FNM"})
	var meta GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&meta)
	resp.Body.Close()

	// No Apply has fired yet — replay is empty → 204.
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/replay", adminSess.Token)
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("empty replay status: got %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	resp.Body.Close()

	// Seat Alice through the HTTP API so the test holds a real
	// player session for the replay-gating assertions below; Bob
	// joins via the lobby directly (no session needed, and it keeps
	// the rate-limited request count under the bucket's burst).
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("join Alice status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	var aliceSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&aliceSess)
	resp.Body.Close()
	alice := aliceSess.PlayerID
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("Join Bob: %v", err)
	}
	// 20+ fillers so the opening hand (7) + post-start draws don't
	// bankrupt the library mid-test.
	aliceDeck := []game.Card{game.NewCommander("A", uuid.Nil)}
	bobDeck := []game.Card{game.NewCommander("B", uuid.Nil)}
	for i := 0; i < 20; i++ {
		aliceDeck = append(aliceDeck, game.NewCard(fmt.Sprintf("AF%d", i), uuid.Nil))
		bobDeck = append(bobDeck, game.NewCard(fmt.Sprintf("BF%d", i), uuid.Nil))
	}
	if _, err := l.SetDeck(meta.ID, alice, "dummy", aliceDeck); err != nil {
		t.Fatalf("SetDeck Alice: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, bob, "dummy", bobDeck); err != nil {
		t.Fatalf("SetDeck Bob: %v", err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	// Fire an Apply by connecting a WS and sending draw_card. Each
	// successful action produces exactly one replay line.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws?token=" + adminSess.Token + "&game=" + meta.ID.String() + "&player=" + alice.String()
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("ws dial: %v", err)
	}
	defer conn.Close()

	// Consume initial snapshot (no replay append — Snapshot path).
	_, _, _ = conn.ReadMessage()

	// One action → one replay line.
	actionID := "00000000-0000-4000-8000-000000000001"
	frame := []byte(`{"v":0,"kind":"action","id":"` + actionID + `","payload":{"type":"draw_card","player":"` + alice.String() + `"}}`)
	if err := conn.WriteMessage(websocket.TextMessage, frame); err != nil {
		t.Fatalf("write action: %v", err)
	}
	_, respRaw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read action response: %v", err)
	}
	// Surface the response so action rejects (bad state, caller
	// mismatch, etc.) don't just look like "no replay file yet".
	var respFrame struct {
		Kind    string          `json:"kind"`
		Payload json.RawMessage `json:"payload"`
	}
	_ = json.Unmarshal(respRaw, &respFrame)
	if respFrame.Kind != "snapshot" {
		t.Fatalf("expected snapshot after draw_card, got %q (payload=%s)",
			respFrame.Kind, respFrame.Payload)
	}

	// Download the replay. Admins may pull it at any time, including
	// mid-game.
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/replay", adminSess.Token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("replay status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); ct != "application/x-ndjson" {
		t.Errorf("content type: got %q, want %q", ct, "application/x-ndjson")
	}
	lines := strings.Split(strings.TrimSuffix(string(body), "\n"), "\n")
	// Lobby-phase mutations route through Room.ApplyExternal, so the
	// replay records the full game setup too: 2 joins + 2 deck
	// uploads + start, then the draw_card action.
	if len(lines) != 6 {
		t.Errorf("replay line count after setup + one action: got %d, want 6\n---\n%s", len(lines), body)
	}

	// Mid-game, a seated player must NOT get the replay — the JSONL
	// carries the unfiltered view (opponents' hands + library order),
	// which is live hidden information while the game runs.
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/replay", aliceSess.Token)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("mid-game player replay status: got %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	resp.Body.Close()

	// Unauth request (no token) → 401.
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/replay", "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("unauth replay status: got %d, want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	resp.Body.Close()

	// Cross-game player → 403. Create a second game, get a player
	// session bound to it; they shouldn't be able to download the
	// first game's replay.
	resp = postJSON(t, srv, "/games", adminSess.Token, createGameRequest{Name: "Other"})
	var other GameMeta
	_ = json.NewDecoder(resp.Body).Decode(&other)
	resp.Body.Close()
	resp = postJSON(t, srv, "/games/"+other.ID.String()+"/join", "",
		joinRequest{InviteToken: other.InviteToken, Name: "Intruder"})
	var intruderSess sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&intruderSess)
	resp.Body.Close()
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/replay", intruderSess.Token)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-game replay status: got %d, want %d", resp.StatusCode, http.StatusForbidden)
	}
	resp.Body.Close()

	// Once the game ends, the seated player can pull their replay.
	g, err := l.LookupGame(meta.ID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	g.End()
	resp = doGet(t, srv, "/games/"+meta.ID.String()+"/replay", aliceSess.Token)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("post-game player replay status: got %d, want %d", resp.StatusCode, http.StatusOK)
	}
	resp.Body.Close()
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
	if _, err := a.Validate(context.TODO(), s.Token); err != nil {
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
	if _, err := a.Validate(context.TODO(), s.Token); err == nil {
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

	// Use an otherwise-valid 100-card source so the 403 we assert is
	// coming from the auth check, not from parsing or validation
	// failing first. If a future refactor reorders the handler, this
	// test won't pass trivially.
	valid := "Commander:\n1 Test Commander\nMainboard:\n99 Plains\n"
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", carolSession.Token,
		uploadDeckRequest{Format: "text", Source: valid, PlayerID: alice})
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("cross-player upload: got %d, want 403", resp.StatusCode)
	}
	resp.Body.Close()

	// Sanity: Carol can upload her OWN deck with the same source. This
	// anchors the 403 above to the player-id mismatch rather than
	// something stateful about the deck bytes.
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", carolSession.Token,
		uploadDeckRequest{Format: "text", Source: valid, PlayerID: carolSession.PlayerID})
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("own upload: got %d, want 200 (body=%s)", resp.StatusCode, body)
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

// TestUploadDeckUnknownCardReturnsViolations verifies the 422 body
// for unknown cards uses the same `{"error", "violations"}` shape as
// validation failures. Regression for the pre-fix behaviour where
// Resolve-time failures returned a bare `{"error": "..."}` string
// that the client couldn't render row-by-row.
func TestUploadDeckUnknownCardReturnsViolations(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)
	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	source := "Commander:\n1 Test Commander\nMainboard:\n1 Not A Real Card\n"
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: source, PlayerID: joined.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unknown-card upload: got %d, want 422", resp.StatusCode)
	}
	var body struct {
		Error      string `json:"error"`
		Violations []struct {
			Code string `json:"code"`
			Card string `json:"card"`
		} `json:"violations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Violations) != 1 {
		t.Fatalf("violations: got %d, want 1 (%+v)", len(body.Violations), body)
	}
	if body.Violations[0].Code != "unknown_card" {
		t.Errorf("code: got %q, want unknown_card", body.Violations[0].Code)
	}
	if body.Violations[0].Card != "Not A Real Card" {
		t.Errorf("card: got %q, want \"Not A Real Card\"", body.Violations[0].Card)
	}
}

// TestUploadDeckBodySizeCap verifies the MaxBytesReader on the deck
// endpoint. A 4 MiB payload must be rejected with 413 before the
// parser ever sees it.
func TestUploadDeckBodySizeCap(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l, _, _ := newTestHTTPStackWithCards(t, idx)
	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	huge := strings.Repeat("x", 4*1024*1024)
	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{Format: "text", Source: huge, PlayerID: joined.PlayerID})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized upload: got %d, want 413", resp.StatusCode)
	}
}

// TestStartReturns409WhenDeckMissing verifies that ErrDeckNotUploaded
// maps to 409 over the HTTP surface. Regression for the pre-fix
// behaviour where the error fell through writeLobbyError's switch
// and surfaced as 500.
func TestStartReturns409WhenDeckMissing(t *testing.T) {
	srv, l, _ := newTestHTTPStack(t)
	meta, _ := l.Create("FNM")

	// Two joins → two seats, neither with an uploaded deck.
	var aliceToken string
	for _, name := range []string{"Alice", "Bob"} {
		r := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
			joinRequest{InviteToken: meta.InviteToken, Name: name})
		var s sessionResponse
		_ = json.NewDecoder(r.Body).Decode(&s)
		r.Body.Close()
		if name == "Alice" {
			aliceToken = s.Token
		}
	}

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/start", aliceToken, nil)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("start without decks: got %d, want 409", resp.StatusCode)
	}
}

// TestUploadDeckViaMoxfieldURL covers the S06.5 URL-import happy
// path end-to-end: POST /games/{id}/decks with format: "url" and a
// Moxfield deck URL. The upstream is stubbed via httptest.Server +
// deck.TestingSetMoxfieldAPIHost, so the handler exercises the full
// fetch → parse → resolve → validate → SetDeck pipeline without
// touching the live API.
func TestUploadDeckViaMoxfieldURL(t *testing.T) {
	idx := buildMinimalDeckIndex(t)

	moxStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v3/decks/all/xyz789" {
			t.Errorf("upstream path: got %q, want /v3/decks/all/xyz789", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		// Matches buildMinimalDeckIndex: 1 commander + 99 Plains.
		body := `{
			"name": "URL Test Deck",
			"boards": {
				"commanders": { "count": 1, "cards": {
					"c1": { "quantity": 1, "card": { "name": "Test Commander" } }
				}},
				"mainboard":  { "count": 99, "cards": {
					"m1": { "quantity": 99, "card": { "name": "Plains" } }
				}}
			}
		}`
		fmt.Fprint(w, body)
	}))
	t.Cleanup(moxStub.Close)
	deck.TestingSetMoxfieldAPIHost(t, moxStub.URL)

	srv, l := newTestHTTPStackFull(t, idx, moxStub.Client())
	meta, _ := l.Create("FNM")
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(resp.Body).Decode(&joined)
	resp.Body.Close()

	resp = postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{
			Format:   "url",
			Source:   "https://moxfield.com/decks/xyz789",
			PlayerID: joined.PlayerID,
		})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("url upload: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
	var ok uploadDeckResponse
	if err := json.NewDecoder(resp.Body).Decode(&ok); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ok.DeckName != "URL Test Deck" {
		t.Errorf("deck_name: got %q, want %q", ok.DeckName, "URL Test Deck")
	}
	if ok.CardCount != 100 {
		t.Errorf("card_count: got %d, want 100", ok.CardCount)
	}
}

// TestUploadDeckURLAutoDetect verifies that omitting `format` with a
// source that starts with https:// lands in the URL path, not the
// text parser (which would see an unknown-name for "https").
func TestUploadDeckURLAutoDetect(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	moxStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"name":"auto","boards":{"commanders":{"count":1,"cards":{"c1":{"quantity":1,"card":{"name":"Test Commander"}}}},"mainboard":{"count":99,"cards":{"m1":{"quantity":99,"card":{"name":"Plains"}}}}}}`)
	}))
	t.Cleanup(moxStub.Close)
	deck.TestingSetMoxfieldAPIHost(t, moxStub.URL)

	srv, l := newTestHTTPStackFull(t, idx, moxStub.Client())
	meta, _ := l.Create("FNM")
	r := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(r.Body).Decode(&joined)
	r.Body.Close()

	// No Format — handler auto-detects from the leading https://.
	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{
			Source:   "https://moxfield.com/decks/auto123",
			PlayerID: joined.PlayerID,
		})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("auto-detect url upload: got %d, want 200 (body=%s)", resp.StatusCode, body)
	}
}

// TestUploadDeckURLUnknownSource covers the 422 shape for a URL
// whose host isn't a supported deck-builder. The response should
// carry a typed `unknown_source` violation rather than a raw 400.
func TestUploadDeckURLUnknownSource(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	srv, l := newTestHTTPStackFull(t, idx, nil)
	meta, _ := l.Create("FNM")
	r := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(r.Body).Decode(&joined)
	r.Body.Close()

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{
			Format:   "url",
			Source:   "https://tappedout.net/mtg-decks/some-deck/",
			PlayerID: joined.PlayerID,
		})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("unknown-source url: got %d, want 422", resp.StatusCode)
	}
	var body struct {
		Violations []struct {
			Code string `json:"code"`
			Card string `json:"card"`
		} `json:"violations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Violations) != 1 || body.Violations[0].Code != "unknown_source" {
		t.Errorf("violations: got %+v, want one unknown_source", body.Violations)
	}
	if body.Violations[0].Card != "https://tappedout.net/mtg-decks/some-deck/" {
		t.Errorf("echo: got %q, want the URL echoed", body.Violations[0].Card)
	}
}

// TestUploadDeckURLNotFound maps upstream 404 → structured
// `deck_not_found` violation.
func TestUploadDeckURLNotFound(t *testing.T) {
	idx := buildMinimalDeckIndex(t)
	moxStub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	t.Cleanup(moxStub.Close)
	deck.TestingSetMoxfieldAPIHost(t, moxStub.URL)

	srv, l := newTestHTTPStackFull(t, idx, moxStub.Client())
	meta, _ := l.Create("FNM")
	r := postJSON(t, srv, "/games/"+meta.ID.String()+"/join", "",
		joinRequest{InviteToken: meta.InviteToken, Name: "Alice"})
	var joined sessionResponse
	_ = json.NewDecoder(r.Body).Decode(&joined)
	r.Body.Close()

	resp := postJSON(t, srv, "/games/"+meta.ID.String()+"/decks", joined.Token,
		uploadDeckRequest{
			Format:   "url",
			Source:   "https://moxfield.com/decks/missing",
			PlayerID: joined.PlayerID,
		})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("deck-not-found: got %d, want 422", resp.StatusCode)
	}
	var body struct {
		Violations []struct{ Code string } `json:"violations"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if len(body.Violations) != 1 || body.Violations[0].Code != "deck_not_found" {
		t.Errorf("violations: got %+v, want one deck_not_found", body.Violations)
	}
}

// newTestHTTPStackFull is a superset of newTestHTTPStackWithCards that
// also wires a DeckHTTPClient. Used by the S06.5 URL-import tests so
// the upload path hits an httptest stub instead of the live Moxfield
// / Archidekt APIs. Passing a nil client preserves the existing
// default-client behaviour.
func newTestHTTPStackFull(t *testing.T, idx *cards.Index, deckClient *http.Client) (*httptest.Server, *Lobby) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})

	cfg := Config{
		Lobby:          l,
		Auth:           a,
		AdminToken:     "shared-admin-token",
		Cards:          idx,
		DeckHTTPClient: deckClient,
	}
	mux := http.NewServeMux()
	mux.Handle("/", Handler(cfg))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, l
}

// TestDevRelaxRateLimitsKnob proves CMDCTRL_DEV_RELAX_RATE_LIMITS
// makes the credential-bearing limiters effectively unlimited: a
// burst far past the default 5-token bucket must not 429. The knob
// is read at Handler construction, so the env var is set before the
// stack is built.
func TestDevRelaxRateLimitsKnob(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	srv, _, _ := newTestHTTPStack(t)

	for i := 0; i < 25; i++ {
		resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
		if resp.StatusCode == http.StatusTooManyRequests {
			t.Fatalf("request %d throttled despite CMDCTRL_DEV_RELAX_RATE_LIMITS", i)
		}
		resp.Body.Close()
	}
}

// TestRateLimitDefaultStillThrottles is the inverse guard: without
// the dev knob, the login bucket still throttles past its burst.
func TestRateLimitDefaultStillThrottles(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "")
	srv, _, _ := newTestHTTPStack(t)

	throttled := false
	for i := 0; i < 8; i++ {
		resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "wrong-token-aaaa"})
		if resp.StatusCode == http.StatusTooManyRequests {
			throttled = true
		}
		resp.Body.Close()
	}
	if !throttled {
		t.Error("8 rapid logins never hit 429; default limiter not enforced")
	}
}

// TestSecureCookiesDecision covers the CMDCTRL_SECURE_COOKIES
// override and the https-redirect-URI default.
func TestSecureCookiesDecision(t *testing.T) {
	httpsCfg := Config{Discord: discord.Config{RedirectURI: "https://cmd.example.io/auth/discord/callback"}}
	httpCfg := Config{Discord: discord.Config{RedirectURI: "http://localhost:8080/auth/discord/callback"}}
	bareCfg := Config{}

	// Unset (blank counts as unset): follow the redirect-URI scheme.
	t.Setenv("CMDCTRL_SECURE_COOKIES", "")
	if !secureCookies(httpsCfg) {
		t.Error("https redirect URI: want Secure by default")
	}
	if secureCookies(httpCfg) {
		t.Error("http redirect URI: want Secure off by default")
	}
	if secureCookies(bareCfg) {
		t.Error("no redirect URI (local dev): want Secure off by default")
	}

	// Explicit truthy forces Secure on even for local http.
	t.Setenv("CMDCTRL_SECURE_COOKIES", "1")
	if !secureCookies(bareCfg) {
		t.Error("CMDCTRL_SECURE_COOKIES=1: want Secure on")
	}
	// Explicit falsy forces Secure off even behind https.
	t.Setenv("CMDCTRL_SECURE_COOKIES", "false")
	if secureCookies(httpsCfg) {
		t.Error("CMDCTRL_SECURE_COOKIES=false: want Secure off")
	}
}

// TestSessionCookieSecureAttribute end-to-ends the flag: with the
// override on, the Set-Cookie from /admin/login must carry Secure.
func TestSessionCookieSecureAttribute(t *testing.T) {
	t.Setenv("CMDCTRL_SECURE_COOKIES", "1")
	srv, _, _ := newTestHTTPStack(t)
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: "shared-admin-token"})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login status: got %d, want 200", resp.StatusCode)
	}
	var found bool
	for _, ck := range resp.Cookies() {
		if ck.Name == auth.SessionCookie {
			found = true
			if !ck.Secure {
				t.Error("session cookie missing Secure attribute with CMDCTRL_SECURE_COOKIES=1")
			}
		}
	}
	if !found {
		t.Fatal("no session cookie set by /admin/login")
	}
}

// TestDecodeJSONBodyCap proves decodeJSON rejects oversized bodies
// with 413 instead of streaming them into the decoder.
func TestDecodeJSONBodyCap(t *testing.T) {
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1") // keep the limiter out of the way
	srv, _, _ := newTestHTTPStack(t)

	huge := strings.Repeat("x", (64<<10)+1024)
	resp := postJSON(t, srv, "/admin/login", "", adminLoginRequest{Token: huge})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body status: got %d, want %d", resp.StatusCode, http.StatusRequestEntityTooLarge)
	}
}
