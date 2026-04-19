package ws

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// newMultiGameServer stands up a hub with a RoomManager holding N
// rooms (each with a 2-seat deterministic game). Returns the ws URL,
// the slice of rooms (for seat ID lookups), and a cleanup function.
func newMultiGameServer(t *testing.T, n int) (string, []*Room, func()) {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	mgr := NewRoomManager(log, "")
	rooms := make([]*Room, 0, n)
	for i := range n {
		g := buildTestGame(t, string(rune('A'+i)))
		rooms = append(rooms, mgr.Create(g))
	}
	hub := NewHub(log)
	hub.SetManager(mgr)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ws", hub.ServeWS)
	srv := httptest.NewServer(mux)
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	return wsURL, rooms, srv.Close
}

// TestHubBindsClientToRequestedGame verifies that a client asking for
// game B never receives broadcasts from actions in game A. This is
// the core multi-game isolation property that the RoomManager and
// broadcastToRoom together must guarantee.
func TestHubBindsClientToRequestedGame(t *testing.T) {
	wsURL, rooms, cleanup := newMultiGameServer(t, 2)
	defer cleanup()

	roomA, roomB := rooms[0], rooms[1]
	seatA0 := roomA.Game.Seats[0].ID
	seatB0 := roomB.Game.Seats[0].ID

	connA := dial(t, wsURL+"?game="+roomA.Game.ID.String()+"&player="+seatA0.String())
	defer connA.Close()
	connB := dial(t, wsURL+"?game="+roomB.Game.ID.String()+"&player="+seatB0.String())
	defer connB.Close()

	// Drain each client's initial snapshot.
	initA := readSnapshotFrame(t, connA)
	initB := readSnapshotFrame(t, connB)
	if initA.Game.ID == initB.Game.ID {
		t.Fatalf("both clients saw the same game id; binding failed")
	}
	if initA.Game.ID != roomA.Game.ID.String() {
		t.Errorf("connA initial: got game %q, want %q", initA.Game.ID, roomA.Game.ID)
	}
	if initB.Game.ID != roomB.Game.ID.String() {
		t.Errorf("connB initial: got game %q, want %q", initB.Game.ID, roomB.Game.ID)
	}

	// Action in game A must NOT broadcast to connB.
	sendActionFrame(t, connA, protocol.ActionPayload{
		Type:   "draw_card",
		Player: seatA0.String(),
	})
	afterA := readSnapshotFrame(t, connA)
	if afterA.Game.ID != roomA.Game.ID.String() {
		t.Errorf("post-action frame from wrong game: got %q, want %q", afterA.Game.ID, roomA.Game.ID)
	}
	// Opening hand of 7 dealt at Start; seat A's draw makes it 8.
	if afterA.Game.Seats[0].Hand.Count != 8 {
		t.Errorf("connA did not see its own draw: hand=%d, want 8", afterA.Game.Seats[0].Hand.Count)
	}

	// connB must see no frame from A's action. Short deadline read.
	if leaked := tryReadSnapshotFrame(t, connB, 200); leaked != nil {
		t.Errorf("connB leaked a frame from game A: %+v", leaked)
	}
}

// TestHubFiltersHandForViewer verifies the per-client visibility
// filter: seat 0 sees its own hand but seat 1's hand is hidden, and
// vice versa, across the same broadcast.
func TestHubFiltersHandForViewer(t *testing.T) {
	wsURL, rooms, cleanup := newMultiGameServer(t, 1)
	defer cleanup()

	room := rooms[0]
	seat0 := room.Game.Seats[0].ID
	seat1 := room.Game.Seats[1].ID

	conn0 := dial(t, wsURL+"?game="+room.Game.ID.String()+"&player="+seat0.String())
	defer conn0.Close()
	conn1 := dial(t, wsURL+"?game="+room.Game.ID.String()+"&player="+seat1.String())
	defer conn1.Close()

	readSnapshotFrame(t, conn0) // initial for conn0
	readSnapshotFrame(t, conn1) // initial for conn1

	// Seat 0 draws. Both clients receive the broadcast, but with
	// different filters applied.
	sendActionFrame(t, conn0, protocol.ActionPayload{
		Type:   "draw_card",
		Player: seat0.String(),
	})
	snap0 := readSnapshotFrame(t, conn0)
	snap1 := readSnapshotFrame(t, conn1)

	// Opening hand of 7 (S08) plus seat 0's draw = 8.
	// From conn0's perspective: seat 0 hand visible, seat 1 hand hidden.
	if len(snap0.Game.Seats[0].Hand.Cards) != 8 {
		t.Errorf("conn0 own hand: got %d cards, want 8", len(snap0.Game.Seats[0].Hand.Cards))
	}
	if len(snap0.Game.Seats[1].Hand.Cards) != 0 {
		t.Errorf("conn0 opponent hand: should be hidden, got %d cards", len(snap0.Game.Seats[1].Hand.Cards))
	}

	// From conn1's perspective: seat 0 hand hidden (count survives),
	// seat 1 hand visible.
	if snap1.Game.Seats[0].Hand.Count != 8 {
		t.Errorf("conn1 opponent hand count: got %d, want 8 (count must survive filter)", snap1.Game.Seats[0].Hand.Count)
	}
	if len(snap1.Game.Seats[0].Hand.Cards) != 0 {
		t.Errorf("conn1 opponent hand cards: should be hidden, got %d", len(snap1.Game.Seats[0].Hand.Cards))
	}
	// conn1 sees its own opening hand of 7 (it hasn't drawn yet).
	if snap1.Game.Seats[1].Hand.Count != 7 {
		t.Errorf("conn1 own hand count: got %d, want 7 (opening hand)", snap1.Game.Seats[1].Hand.Count)
	}

	// Seq equality across clients is still enforced — both snapshots
	// are views of the SAME captured state.
	if snap0.Seq != snap1.Seq {
		t.Errorf("seq mismatch across viewers: conn0=%d conn1=%d", snap0.Seq, snap1.Seq)
	}
}

// TestHubRejectsUnknownGame verifies the upgrade-path reject when a
// client asks for a game that is not registered.
func TestHubRejectsUnknownGame(t *testing.T) {
	wsURL, _, cleanup := newMultiGameServer(t, 1)
	defer cleanup()

	// A random UUID that is guaranteed not to match any registered
	// room.
	bogus := uuid.New().String()
	u := wsURL + "?game=" + bogus
	conn, resp, err := tryDial(u)
	if conn != nil {
		conn.Close()
		t.Fatal("dial with bogus game id unexpectedly succeeded")
	}
	if err == nil {
		t.Fatal("expected dial error, got nil")
	}
	if resp == nil {
		t.Fatal("expected HTTP response on failed upgrade, got nil")
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

// TestHubRejectsPlayerNotInGame verifies that a player ID from one
// game cannot be used to bind into another game's view. A regression
// here would let a malicious client claim a seat that doesn't exist
// and silently receive spectator frames.
func TestHubRejectsPlayerNotInGame(t *testing.T) {
	wsURL, rooms, cleanup := newMultiGameServer(t, 2)
	defer cleanup()

	gameA := rooms[0].Game.ID
	seatFromB := rooms[1].Game.Seats[0].ID

	u := wsURL + "?game=" + gameA.String() + "&player=" + seatFromB.String()
	conn, resp, err := tryDial(u)
	if conn != nil {
		conn.Close()
		t.Fatal("dial with mismatched player/game unexpectedly succeeded")
	}
	if err == nil {
		t.Fatal("expected dial error, got nil")
	}
	if resp == nil || resp.StatusCode != http.StatusForbidden {
		t.Errorf("status: got %v, want %d", resp, http.StatusForbidden)
	}
}
