package lobby

// seat_remove_started_test.go: #3093, an admin removes a seat from a
// STARTED game. The removal is a concession (CR 800.4a), and the seat's
// session is refused afterwards the way #2785's is.

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// startedTable seats n named humans (the first is the host), gives each
// a deck and starts the game. creator owns the table.
func (s *seatRemoveStack) startedTable(t *testing.T, creator uuid.UUID, names ...string) (GameMeta, []sessionResponse) {
	t.Helper()
	meta, err := s.l.CreateBy("Started table", creator)
	if err != nil {
		t.Fatal(err)
	}
	var seats []sessionResponse
	for _, n := range names {
		seats = append(seats, joinAs(t, s.srv, meta, n))
	}
	for _, p := range seats {
		if _, err := s.l.SetDeck(meta.ID, p.PlayerID, "deck", botDeck(30)); err != nil {
			t.Fatalf("SetDeck: %v", err)
		}
	}
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return meta, seats
}

func (s *seatRemoveStack) enginePlayer(t *testing.T, gameID, playerID uuid.UUID) *game.Player {
	t.Helper()
	p := s.l.RoomOf(gameID).Game.PlayerByID(playerID)
	if p == nil {
		t.Fatalf("player %s is not in the engine", playerID)
	}
	return p
}

func TestRemoveStartedSeatAuthMatrix(t *testing.T) {
	s := newSeatRemoveStack(t)
	creator := uuid.New()
	meta, seats := s.startedTable(t, creator, "Alice", "Bob", "Carol")
	alice, bob, carol := seats[0], seats[1], seats[2]
	admin := adminToken(t, s.srv)

	spec := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		map[string]string{"invite_token": meta.SpectatorInvite}), http.StatusOK)

	// Every non-admin is refused, and the table is untouched.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, ""), http.StatusUnauthorized)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, carol.Token), http.StatusForbidden)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, bob.Token), http.StatusForbidden)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, identifiedToken(t, s.a, uuid.New())), http.StatusForbidden)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, spec.Token), http.StatusForbidden)
	// The host and the creator keep only the unstarted-table power: 409.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, alice.Token), http.StatusConflict)
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, identifiedToken(t, s.a, creator)), http.StatusConflict)
	if p := s.enginePlayer(t, meta.ID, bob.PlayerID); p.Eliminated || p.Removed {
		t.Fatal("a refused request removed the seat")
	}
	if _, err := s.a.Validate(context.Background(), bob.Token); err != nil {
		t.Fatalf("Bob's session broke on a refused request: %v", err)
	}

	// The admin's refusals: the host's seat, a stranger id, a bad id.
	s.wantStatus(t, s.del(t, meta.ID, alice.PlayerID, admin), http.StatusUnprocessableEntity)
	s.wantStatus(t, s.del(t, meta.ID, uuid.New(), admin), http.StatusForbidden)
	s.wantStatus(t, s.del(t, uuid.New(), bob.PlayerID, admin), http.StatusNotFound)

	// The admin removes Bob: 200, the seat stays in the list as a
	// player who left, the game goes on.
	resp := s.del(t, meta.ID, bob.PlayerID, admin)
	s.wantStatus(t, resp, http.StatusOK)
	after := mustGet(t, s.l, meta.ID)
	if len(after.Players) != 3 {
		t.Errorf("seats = %d, want 3: a started table keeps the removed seat in its list", len(after.Players))
	}
	if after.State != string(game.StateActive) {
		t.Errorf("state = %s, want active", after.State)
	}
	if p := s.enginePlayer(t, meta.ID, bob.PlayerID); !p.Eliminated || !p.Removed {
		t.Errorf("Bob: eliminated=%v removed=%v, want both", p.Eliminated, p.Removed)
	}
	for _, p := range []sessionResponse{alice, carol} {
		if _, err := s.a.Validate(context.Background(), p.Token); err != nil {
			t.Errorf("%s's session broke: %v", p.Principal.Name, err)
		}
	}
	// Removing the same seat again is fine: the seat is only re-marked.
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, admin), http.StatusOK)
}

func TestRemoveStartedSeatRefusesABotSeat(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, err := s.l.Create("Bot table")
	if err != nil {
		t.Fatal(err)
	}
	alice := joinAs(t, s.srv, meta, "Alice")
	_, botID, err := s.l.AddBot(meta.ID, "Bot", "random", "", "d", botDeck(30))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.l.SetDeck(meta.ID, alice.PlayerID, "deck", botDeck(30)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	s.wantStatus(t, s.del(t, meta.ID, botID, adminToken(t, s.srv)), http.StatusUnprocessableEntity)
	if p := s.enginePlayer(t, meta.ID, botID); p.Eliminated {
		t.Error("the bot seat was conceded through the human-seat route")
	}
}

func TestRemoveStartedSeatEndsATwoPlayerGame(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob")
	alice, bob := seats[0], seats[1]
	admin := adminToken(t, s.srv)

	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, admin), http.StatusOK)
	after := mustGet(t, s.l, meta.ID)
	if after.State != string(game.StateEnded) {
		t.Fatalf("state = %s, want ended: one player is left", after.State)
	}
	seat, ok := s.l.RoomOf(meta.ID).Game.WinnerSeat()
	if !ok || seat != 0 {
		t.Errorf("winner = seat %d (ok=%v), want Alice's seat 0", seat, ok)
	}
	// Alice, the winner, is not signed out. Bob is.
	if _, err := s.a.Validate(context.Background(), alice.Token); err != nil {
		t.Errorf("the winner's session: %v", err)
	}
	if _, err := s.a.Validate(context.Background(), bob.Token); err != auth.ErrSeatRemoved {
		t.Errorf("Bob's session = %v, want ErrSeatRemoved", err)
	}
	// An ended game has nothing left to leave: 409.
	s.wantStatus(t, s.del(t, meta.ID, alice.PlayerID, admin), http.StatusConflict)
}

func TestRemovedStartedSeatCannotReclaimOrReconnect(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob", "Carol")
	bob := seats[1]
	bobConn, _, err := websocket.DefaultDialer.Dial(s.wsURL(bob.Token), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer bobConn.Close()
	_ = bobConn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if _, _, err := bobConn.ReadMessage(); err != nil {
		t.Fatalf("opening snapshot: %v", err)
	}

	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, adminToken(t, s.srv)), http.StatusOK)

	// The live socket closes with the reason the client matches on.
	for {
		_, _, err := bobConn.ReadMessage()
		if err == nil {
			continue
		}
		ce, ok := err.(*websocket.CloseError)
		if !ok || ce.Code != websocket.CloseNormalClosure || ce.Text != ws.SeatRemovedReason {
			t.Fatalf("Bob's socket ended with %v, want 1000 %q", err, ws.SeatRemovedReason)
		}
		break
	}
	// Every authenticated route and a reconnect refuse him.
	for _, path := range []string{"/me", "/games/" + meta.ID.String()} {
		resp := doGet(t, s.srv, path, bob.Token)
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized || !strings.Contains(string(body), auth.SeatRemovedMessage) {
			t.Errorf("GET %s: %d %s, want 401 %q", path, resp.StatusCode, body, auth.SeatRemovedMessage)
		}
	}
	if _, upgrade, err := websocket.DefaultDialer.Dial(s.wsURL(bob.Token), nil); err == nil || upgrade == nil || upgrade.StatusCode != http.StatusUnauthorized {
		t.Fatalf("reconnect: err=%v upgrade=%v, want a 401 at the upgrade", err, upgrade)
	}
	// A reclaim ticket cannot put him back: not minted, not redeemed.
	if _, err := s.l.MintReclaim(meta.ID, bob.PlayerID); err != ErrPlayerNotInGame {
		t.Errorf("MintReclaim for a removed seat = %v, want ErrPlayerNotInGame", err)
	}
}

// TestRedeemReclaimRefusesASeatRemovedAfterTheMint: a ticket minted
// while the seat was in play is dead once an admin removes the seat.
func TestRedeemReclaimRefusesASeatRemovedAfterTheMint(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob", "Carol")
	bob := seats[1]
	ticket, err := s.l.MintReclaim(meta.ID, bob.PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, adminToken(t, s.srv)), http.StatusOK)
	if _, _, err := s.l.RedeemReclaim(meta.ID, ticket.Token); err != ErrPlayerNotInGame {
		t.Errorf("RedeemReclaim = %v, want ErrPlayerNotInGame", err)
	}
}

// TestAConcededPlayerKeepsTheirSession: the line the new flag draws. A
// player who gave up is not a removed player, and their credential may
// keep watching their own table.
func TestAConcededPlayerKeepsTheirSession(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob", "Carol")
	bob := seats[1]
	if err := s.l.RoomOf(meta.ID).Game.Concede(bob.PlayerID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.a.Validate(context.Background(), bob.Token); err != nil {
		t.Errorf("a player who conceded was signed out: %v", err)
	}
	if s.l.SeatRemoved(meta.ID, bob.PlayerID) {
		t.Error("a concession reads as a removal")
	}
}

// TestRemovedStartedSeatSurvivesARestart: the refusal is read from the
// engine's snapshot, so a restarted server still turns the player away.
func TestRemovedStartedSeatSurvivesARestart(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	meta, err := l.Create("Durable")
	if err != nil {
		t.Fatal(err)
	}
	var ids []uuid.UUID
	for _, n := range []string{"Alice", "Bob", "Carol"} {
		_, id, err := l.Join(meta.ID, meta.InviteToken, n)
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, id)
		if _, err := l.SetDeck(meta.ID, id, "deck", botDeck(30)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	bob, carol := ids[1], ids[2]
	if _, err := l.RemoveStartedSeat(meta.ID, bob); err != nil {
		t.Fatal(err)
	}
	if !l.SeatRemoved(meta.ID, bob) {
		t.Fatal("the live lobby does not refuse the removed seat")
	}

	// --- the deploy -----------------------------------------------
	l2, _ := newDurableLobby(t, dir)
	if n := l2.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}
	if !l2.SeatRemoved(meta.ID, bob) {
		t.Error("the restored lobby lets the removed seat back in")
	}
	if l2.SeatRemoved(meta.ID, carol) || l2.SeatRemoved(meta.ID, ids[0]) {
		t.Error("the restored lobby refuses a seat nobody removed")
	}
	// Through the real wrapper, with a token minted before the restart.
	a := newHMAC(t)
	tok, _, err := a.Issue(context.Background(), auth.Principal{
		Role: auth.RolePlayer, GameID: meta.ID, PlayerID: bob,
	}, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.WithSeatCheck(a, l2).Validate(context.Background(), tok); err != auth.ErrSeatRemoved {
		t.Errorf("Validate after the restart = %v, want ErrSeatRemoved", err)
	}
	if _, err := l2.MintReclaim(meta.ID, bob); err != ErrPlayerNotInGame {
		t.Errorf("MintReclaim after the restart = %v, want ErrPlayerNotInGame", err)
	}
}

// TestRemoveStartedSeatIsFinal: the undo ring is emptied, so a seat that
// the table has refused cannot be put back by a later Undo.
func TestRemoveStartedSeatIsFinal(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Undo table")
	var ids []uuid.UUID
	for _, n := range []string{"Alice", "Bob", "Carol"} {
		_, id, _ := l.Join(meta.ID, meta.InviteToken, n)
		ids = append(ids, id)
		if _, err := l.SetDeck(meta.ID, id, "deck", botDeck(30)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	room := l.RoomOf(meta.ID)
	// An ordinary action puts an entry on the undo ring.
	if _, _, err := room.Apply(ids[0], func() error { return nil }); err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if _, err := l.RemoveStartedSeat(meta.ID, ids[1]); err != nil {
		t.Fatal(err)
	}
	if _, _, err := room.Undo(uuid.Nil); !ws.IsErrNothingToUndo(err) {
		t.Fatalf("Undo after a removal = %v, want nothing to undo", err)
	}
	if !room.Game.PlayerByID(ids[1]).Removed {
		t.Error("the removal was undone")
	}
}

func TestRemoveStartedSeatRefusals(t *testing.T) {
	l := newTestLobby(t)
	meta, _ := l.Create("Lobby table")
	_, alice, _ := l.Join(meta.ID, meta.InviteToken, "Alice")
	_, bob, _ := l.Join(meta.ID, meta.InviteToken, "Bob")
	if _, err := l.RemoveStartedSeat(meta.ID, bob); err != ErrGameNotInProgress {
		t.Errorf("an unstarted table: %v, want ErrGameNotInProgress", err)
	}
	if _, err := l.RemoveStartedSeat(uuid.New(), bob); err != ErrGameNotFound {
		t.Errorf("no such table: %v, want ErrGameNotFound", err)
	}
	for _, id := range []uuid.UUID{alice, bob} {
		if _, err := l.SetDeck(meta.ID, id, "deck", botDeck(30)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := l.RemoveStartedSeat(meta.ID, uuid.New()); err != ErrPlayerNotInGame {
		t.Errorf("a stranger id: %v, want ErrPlayerNotInGame", err)
	}
	if _, err := l.RemoveStartedSeat(meta.ID, alice); err != ErrCannotRemoveHost {
		t.Errorf("the host: %v, want ErrCannotRemoveHost", err)
	}
	// An unstarted-table removal still refuses a started table.
	if _, err := l.RemoveSeat(meta.ID, bob); err != ErrGameStarted {
		t.Errorf("RemoveSeat on a started table: %v, want ErrGameStarted", err)
	}
}

// TestRemovedSeatIsNotExemptFromTheConnectionCap: the cap (ADR 0148)
// never asks a seated player, and a removed player is not seated. The
// seat check refuses the credential at authorisation, before the hub
// decides whether the socket counts as a seat's, so the answer at a full
// server is the same 401 as anywhere else and not a "server full" frame
// (which would invite a retry loop) nor an admit.
func TestRemovedSeatIsNotExemptFromTheConnectionCap(t *testing.T) {
	s := newSeatRemoveStack(t)
	meta, seats := s.startedTable(t, uuid.New(), "Alice", "Bob", "Carol")
	alice, bob := seats[0], seats[1]
	s.wantStatus(t, s.del(t, meta.ID, bob.PlayerID, adminToken(t, s.srv)), http.StatusOK)

	// Fill the server: one spectator socket against a cap of one.
	s.hub.SetMaxConnections(1)
	spec := decodeSession(t, postJSON(t, s.srv, "/games/"+meta.ID.String()+"/spectate", "",
		map[string]string{"invite_token": meta.SpectatorInvite}), http.StatusOK)
	specConn, _, err := websocket.DefaultDialer.Dial(s.wsURL(spec.Token), nil)
	if err != nil {
		t.Fatalf("the first connection was refused: %v", err)
	}
	defer specConn.Close()

	_, upgrade, err := websocket.DefaultDialer.Dial(s.wsURL(bob.Token), nil)
	if err == nil || upgrade == nil || upgrade.StatusCode != http.StatusUnauthorized {
		t.Fatalf("removed seat at a full server: err=%v upgrade=%v, want a 401 at the upgrade", err, upgrade)
	}
	// The exemption itself is intact for a seat that is still seated.
	aliceConn, _, err := websocket.DefaultDialer.Dial(s.wsURL(alice.Token), nil)
	if err != nil {
		t.Fatalf("a seated player was refused at the cap: %v", err)
	}
	aliceConn.Close()
}
