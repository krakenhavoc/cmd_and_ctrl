package lobby

// seat_remove_test.go: #2785, the host removes a seat from an
// unstarted table.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// seatRemoveStack is the production wiring for this feature: HMAC
// sessions (so a removed seat's token still carries a good signature,
// which is the case WithSeatCheck exists for), wrapped in the seat
// check, with a real hub as the seat evictor.
type seatRemoveStack struct {
	srv *httptest.Server
	l   *Lobby
	a   auth.Authenticator
	hub *ws.Hub
}

func newSeatRemoveStack(t *testing.T) *seatRemoveStack {
	t.Helper()
	t.Setenv("CMDCTRL_DEV_RELAX_RATE_LIMITS", "1")
	log := quietLogger()
	mgr := ws.NewRoomManager(log, "")
	l := NewLobby(mgr)
	a := auth.WithSeatCheck(newHMAC(t), l)
	hub := ws.NewHub(log)
	hub.SetManager(mgr)
	hub.SetAuthorizer(&WSAuthorizer{Auth: a})
	l.SetStateBroadcaster(hub)
	mux := http.NewServeMux()
	mux.Handle("/", Handler(Config{
		Lobby:       l,
		Auth:        a,
		AdminToken:  "shared-admin-token",
		SeatEvictor: hub,
	}))
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &seatRemoveStack{srv: srv, l: l, a: a, hub: hub}
}

func (s *seatRemoveStack) del(t *testing.T, gameID, playerID uuid.UUID, token string) *http.Response {
	t.Helper()
	return doJSON(t, s.srv, http.MethodDelete, "/games/"+gameID.String()+"/seats/"+playerID.String(), token, nil)
}

func (s *seatRemoveStack) wantStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	defer resp.Body.Close()
	if resp.StatusCode != want {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("status %d, want %d: %s", resp.StatusCode, want, body)
	}
}

// table seats Alice (host), Bob and Carol, and returns each session.
func (s *seatRemoveStack) table(t *testing.T) (meta GameMeta, alice, bob sessionResponse) {
	t.Helper()
	meta, err := s.l.Create("FNM")
	if err != nil {
		t.Fatal(err)
	}
	alice = joinAs(t, s.srv, meta, "Alice")
	bob = joinAs(t, s.srv, meta, "Bob")
	return meta, alice, bob
}

func (s *seatRemoveStack) wsURL(token string) string {
	return "ws" + strings.TrimPrefix(s.srv.URL, "http") + "/ws?token=" + token
}

func TestRemoveSeatAuthMatrix(t *testing.T) {
	s := newSeatRemoveStack(t)
	creator := uuid.New()
	meta, err := s.l.CreateBy("Creator's table", creator)
	if err != nil {
		t.Fatal(err)
	}
	alice := joinAs(t, s.srv, meta, "Alice") // first human: host
	bob := joinAs(t, s.srv, meta, "Bob")
	carol := joinAs(t, s.srv, meta, "Carol")
	dave := joinAs(t, s.srv, meta, "Dave")
	if got := mustGet(t, s.l, meta.ID).HostPlayerID; got != alice.PlayerID {
		t.Fatalf("host = %v, want Alice", got)
	}

	// Anonymous: 401.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, ""), http.StatusUnauthorized)
	// Another player (not the host): 403, and Bob is still seated.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, carol.Token), http.StatusForbidden)
	// A player removing themselves is not the host either: 403.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, bob.Token), http.StatusForbidden)
	// A signed-in stranger (not creator, no seat): 403.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, identifiedToken(t, s.a, uuid.New())), http.StatusForbidden)
	// A spectator: 403.
	spec := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		map[string]string{"invite_token": meta.SpectatorInvite}), http.StatusOK)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, spec.Token), http.StatusForbidden)
	if n := len(mustGet(t, s.l, meta.ID).Players); n != 4 {
		t.Fatalf("a refused request changed the table: %d seats", n)
	}

	// Host: OK.
	resp := s.del(t, meta.ID, bob.PlayerID, alice.Token)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("host remove: %d %s", resp.StatusCode, body)
	}
	var after GameMeta
	if err := json.NewDecoder(resp.Body).Decode(&after); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if len(after.Players) != 3 {
		t.Fatalf("response lists %d seats, want 3", len(after.Players))
	}
	for i, p := range after.Players {
		if p.PlayerID == bob.PlayerID {
			t.Errorf("Bob is still listed: %+v", after.Players)
		}
		if p.Seat != i {
			t.Errorf("seat %d numbered %d: the gap did not close", i, p.Seat)
		}
	}

	// Creator (signed in, never seated): OK.
	s.wantStatus(t, s.del(t, meta.ID, carol.PlayerID, identifiedToken(t, s.a, creator)), http.StatusOK)

	// Admin: OK.
	s.wantStatus(t, s.del(t, meta.ID, dave.PlayerID, adminToken(t, s.srv)), http.StatusOK)
	if got := mustGet(t, s.l, meta.ID).Players; len(got) != 1 || got[0].PlayerID != alice.PlayerID {
		t.Fatalf("seats = %+v, want just Alice", got)
	}
}

func TestRemoveSeatRefusals(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)
	admin := adminToken(t, s.srv)
	_, botID, err := s.l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(5))
	if err != nil {
		t.Fatal(err)
	}

	// The host's own seat can't be removed, not even by the admin.
	s.wantStatus(t, s.del(t, meta.ID, alice.PlayerID, alice.Token), http.StatusUnprocessableEntity)
	s.wantStatus(t, s.del(t, meta.ID, alice.PlayerID, admin), http.StatusUnprocessableEntity)
	// A bot has its own route.
	s.wantStatus(t, s.del(t, meta.ID, botID, admin), http.StatusUnprocessableEntity)
	// A player id that is not a seat here.
	s.wantStatus(t, s.del(t, meta.ID, uuid.New(), admin), http.StatusForbidden)
	// Not a UUID.
	resp := doJSON(t, s.srv, http.MethodDelete, "/games/"+meta.ID.String()+"/seats/nope", admin, nil)
	s.wantStatus(t, resp, http.StatusBadRequest)
	// No such table.
	s.wantStatus(t, s.del(t, uuid.New(), bob.PlayerID, admin), http.StatusNotFound)

	if n := len(mustGet(t, s.l, meta.ID).Players); n != 3 {
		t.Fatalf("a refusal changed the table: %d seats, want 3", n)
	}
}

func TestRemoveSeatRefusesAStartedTable(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)
	uploadDummyDeck(t, s.l, meta.ID, alice.PlayerID)
	uploadDummyDeck(t, s.l, meta.ID, bob.PlayerID)
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// The host is refused on a started table. Only an admin may remove
	// a seat from a started game (#3093); the admin's rows are in
	// seat_remove_started_test.go.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusConflict)
	if n := len(mustGet(t, s.l, meta.ID).Players); n != 2 {
		t.Fatalf("a refused removal changed a started table: %d seats", n)
	}
	if _, err := s.a.Validate(context.Background(), bob.Token); err != nil {
		t.Errorf("Bob's session stopped validating on a refused removal: %v", err)
	}
}

func TestRemovedSeatCannotReclaimOrReconnect(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)
	admin := adminToken(t, s.srv)

	// Before: the token is good everywhere.
	if _, err := s.a.Validate(context.Background(), bob.Token); err != nil {
		t.Fatal(err)
	}
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusOK)

	// The signature is still good; the seat is not.
	if _, err := s.a.Validate(context.Background(), bob.Token); err != auth.ErrSeatRemoved {
		t.Fatalf("Validate after removal = %v, want ErrSeatRemoved", err)
	}
	// Every authenticated route says so, with the message the client
	// matches on. The seat-management route is the one that matters
	// most: a removed seat must not be able to add or remove bots.
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/me"},
		{http.MethodGet, "/games/" + meta.ID.String()},
		{http.MethodDelete, "/games/" + meta.ID.String() + "/seats/bot/" + uuid.NewString()},
		{http.MethodPost, "/games/" + meta.ID.String() + "/seats/bot"},
	} {
		resp := doJSON(t, s.srv, tc.method, tc.path, bob.Token, nil)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(string(body), auth.SeatRemovedMessage) {
			t.Errorf("%s %s with the removed token: %d %s, want 401 %q", tc.method, tc.path, resp.StatusCode, body, auth.SeatRemovedMessage)
		}
	}
	// A reconnect is refused at the upgrade.
	_, upgrade, err := websocket.DefaultDialer.Dial(s.wsURL(bob.Token), nil)
	if err == nil {
		t.Fatal("removed seat reconnected")
	}
	if upgrade == nil || upgrade.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reconnect answered %v, want 401", upgrade)
	}
	// Alice, still seated, and the admin are untouched.
	if _, err := s.a.Validate(context.Background(), alice.Token); err != nil {
		t.Errorf("Alice's session broke: %v", err)
	}
	if resp := doGet(t, s.srv, "/me", admin); resp.StatusCode != http.StatusOK {
		t.Errorf("admin /me: %d", resp.StatusCode)
	} else {
		resp.Body.Close()
	}
}

func TestRemovedSlotIsReusable(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)
	// Fill the table, then free one slot.
	carol := joinAs(t, s.srv, meta, "Carol")
	dave := joinAs(t, s.srv, meta, "Dave")
	resp := postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "", joinRequest{InviteToken: meta.InviteToken, Name: "Eve"})
	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("a full table took a fifth seat: %d", resp.StatusCode)
	}
	resp.Body.Close()

	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusOK)
	eve := joinAs(t, s.srv, meta, "Eve")
	if eve.PlayerID == bob.PlayerID {
		t.Error("the new seat reused the removed player's id; the old token would match it")
	}
	players := mustGet(t, s.l, meta.ID).Players
	if len(players) != 4 {
		t.Fatalf("seats = %d, want 4", len(players))
	}
	// The old credential still does not work now the slot is taken again.
	if _, err := s.a.Validate(context.Background(), bob.Token); err != auth.ErrSeatRemoved {
		t.Errorf("Bob's old token after his slot was refilled: %v", err)
	}
	for _, p := range []sessionResponse{alice, carol, dave, eve} {
		if _, err := s.a.Validate(context.Background(), p.Token); err != nil {
			t.Errorf("%s's token: %v", p.Principal.Name, err)
		}
	}
}

func TestRemoveSeatClosesTheSocketWithTheMessage(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, bob := s.table(t)

	dial := func(token string) *websocket.Conn {
		conn, _, err := websocket.DefaultDialer.Dial(s.wsURL(token), nil)
		if err != nil {
			t.Fatalf("dial: %v", err)
		}
		t.Cleanup(func() { conn.Close() })
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		if _, _, err := conn.ReadMessage(); err != nil { // the opening snapshot
			t.Fatalf("snapshot: %v", err)
		}
		return conn
	}
	bobConn := dial(bob.Token)
	aliceConn := dial(alice.Token)

	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusOK)

	// Bob: reads until the close, and the close names why.
	for {
		_, _, err := bobConn.ReadMessage()
		if err == nil {
			continue // a broadcast that raced the close
		}
		ce, ok := err.(*websocket.CloseError)
		if !ok {
			t.Fatalf("Bob's socket ended with %v, want a close frame", err)
		}
		if ce.Code != websocket.CloseNormalClosure || ce.Text != ws.SeatRemovedReason {
			t.Fatalf("close = %d %q, want 1000 %q", ce.Code, ce.Text, ws.SeatRemovedReason)
		}
		break
	}

	// Alice's socket is still up and hears the new table.
	_, raw, err := aliceConn.ReadMessage()
	if err != nil {
		t.Fatalf("Alice's socket closed: %v", err)
	}
	var f struct {
		Kind string `json:"kind"`
	}
	_ = json.Unmarshal(raw, &f)
	if f.Kind != "snapshot" {
		t.Errorf("Alice got a %q frame, want a snapshot of the smaller table", f.Kind)
	}
}

func TestRemoveSeatRewritesTheStoredSeats(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, _ := l.Create("FNM")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	_, _, _ = l.Join(meta.ID, meta.InviteToken, "Carol")
	if _, err := l.RemoveSeat(meta.ID, bob); err != nil {
		t.Fatal(err)
	}

	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	back, err := l2.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Players) != 2 || back.Players[0].Name != "Alice" || back.Players[1].Name != "Carol" {
		t.Fatalf("restored seats = %+v, want Alice then Carol", back.Players)
	}
	if back.Players[1].Seat != 1 {
		t.Errorf("Carol's stored seat = %d, want 1", back.Players[1].Seat)
	}
	if l2.SeatRemoved(meta.ID, bob) != true {
		t.Error("the restored lobby does not know Bob's seat is gone")
	}
}

func TestRemoveAgentSeat(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, alice, _ := s.table(t)
	agent := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/join", "",
		agentJoinBody(meta.InviteToken, "Claude", "Claude Code")), http.StatusOK)
	s.wantStatus(t, s.del(t, meta.ID, agent.PlayerID, alice.Token), http.StatusOK)
	if _, err := s.a.Validate(context.Background(), agent.Token); err != auth.ErrSeatRemoved {
		t.Errorf("agent token after removal: %v", err)
	}
}

func TestSeatRemovedIsFalseForWhatItDoesNotKnow(t *testing.T) {
	l := newTestLobby(t)
	if l.SeatRemoved(uuid.New(), uuid.New()) {
		t.Error("an unknown game reads as a removed seat")
	}
	meta, _ := l.Create("FNM")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	if l.SeatRemoved(meta.ID, alice) {
		t.Error("a seated player reads as removed")
	}
}

func TestCanRemoveSeat(t *testing.T) {
	game1, game2 := uuid.New(), uuid.New()
	host, other, creator := uuid.New(), uuid.New(), uuid.New()
	meta := GameMeta{ID: game1, HostPlayerID: host, CreatedBy: creator}
	cases := []struct {
		name  string
		p     auth.Principal
		admin bool
		want  bool
	}{
		{"admin", auth.Principal{Role: auth.RoleAdmin}, true, true},
		{"host", auth.Principal{Role: auth.RolePlayer, GameID: game1, PlayerID: host}, false, true},
		{"creator, unseated", auth.Principal{Role: auth.RoleIdentified, UserID: creator}, false, true},
		{"other player", auth.Principal{Role: auth.RolePlayer, GameID: game1, PlayerID: other}, false, false},
		{"host of another game", auth.Principal{Role: auth.RolePlayer, GameID: game2, PlayerID: host}, false, false},
		{"stranger user", auth.Principal{Role: auth.RoleIdentified, UserID: uuid.New()}, false, false},
		{"spectator", auth.Principal{Role: auth.RoleSpectator, GameID: game1}, false, false},
		{"empty", auth.Principal{}, false, false},
	}
	for _, tc := range cases {
		if got := CanRemoveSeat(tc.p, meta, tc.admin); got != tc.want {
			t.Errorf("%s: CanRemoveSeat = %v, want %v", tc.name, got, tc.want)
		}
	}
}
