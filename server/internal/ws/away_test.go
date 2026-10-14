package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0177 §1, §2 and §6, the hub's and the room's halves: presence on
// every frame, the movable mark on an offline seat, the stand-in mark,
// and a stand-in seat's own actions refused.

func seatOf(t *testing.T, snap protocol.SnapshotPayload, id uuid.UUID) protocol.PlayerView {
	t.Helper()
	for _, s := range snap.Game.Seats {
		if s.ID == id.String() {
			return s
		}
	}
	t.Fatalf("seat %s is not in the frame", id)
	return protocol.PlayerView{}
}

func TestPresenceIsStampedOnEveryFrame(t *testing.T) {
	wsURL, g, room, hub, clk := presenceServer(t)
	a, b := g.Seats[0].ID, g.Seats[1].ID
	start := clk.now()

	conn := dialAs(t, wsURL, a)
	defer conn.Close()
	snap := readSnapshotFrame(t, conn)
	// The connecting seat is present in its own first frame, though it
	// was staged before the connection was admitted.
	if p := seatOf(t, snap, a).Presence; p != nil {
		t.Errorf("my own seat in my first frame: presence %+v, want none", p)
	}
	// B has never been connected: offline since the hub first saw it.
	p := seatOf(t, snap, b).Presence
	if p == nil || p.State != protocol.PresenceOffline || p.Since != start.UTC().Format(time.RFC3339) || p.Movable {
		t.Fatalf("a seat nobody connected to: presence %+v, want offline since %v", p, start)
	}

	// The room says the table may move on without B: the mark rides
	// B's presence on the next frame.
	room.SetMovable(map[uuid.UUID]bool{b: true})
	hub.RefreshPresence(g.ID)
	if p := seatOf(t, readSnapshotFrame(t, conn), b).Presence; p == nil || !p.Movable {
		t.Fatalf("a movable offline seat: presence %+v, want movable", p)
	}

	// B connects, then hides its tab: hidden, never movable.
	connB := dialAs(t, wsURL, b)
	defer connB.Close()
	readSnapshotFrame(t, connB)
	hidAt := start.Add(time.Minute)
	clk.set(hidAt)
	sendVisibility(t, connB, `{"visible": false}`)
	awaitAway(t, hub, g.ID, b, true)
	hub.RefreshPresence(g.ID)
	p = seatOf(t, readSnapshotFrame(t, conn), b).Presence
	if p == nil || p.State != protocol.PresenceHidden || p.Movable || p.Since != hidAt.UTC().Format(time.RFC3339) {
		t.Fatalf("a hidden seat: presence %+v, want hidden since %v and not movable", p, hidAt)
	}

	// Closing the tab: offline from the close, not from when it hid.
	closedAt := hidAt.Add(time.Minute)
	clk.set(closedAt)
	_ = connB.Close()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if state, since := hub.SeatPresence(g.ID, b); state == protocol.PresenceOffline {
			if !since.Equal(closedAt) {
				t.Errorf("offline since %v, want the close at %v", since, closedAt)
			}
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the closed seat never read offline")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestABotSeatNeverCarriesPresence(t *testing.T) {
	_, g, room, hub, _ := presenceServer(t)
	g.Seats[1].IsBot = true
	view, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	view = hub.stampPresence(room, view)
	for _, s := range view.Seats {
		if s.ID == g.Seats[1].ID.String() && s.Presence != nil {
			t.Errorf("a bot seat carries presence %+v", s.Presence)
		}
		if s.ID == g.Seats[0].ID.String() && s.Presence == nil {
			t.Error("an unconnected person's seat carries no presence")
		}
	}
}

func TestAStandInSeatIsMarkedAndItsOwnActionsAreRefused(t *testing.T) {
	wsURL, g, room, _, clk := presenceServer(t)
	a := g.Seats[0].ID
	since := clk.now()
	room.SetStandIn(a, "heuristic", since)
	if !room.StandIn(a) {
		t.Fatal("StandIn reads false after SetStandIn")
	}

	conn := dialAs(t, wsURL, a)
	defer conn.Close()
	si := seatOf(t, readSnapshotFrame(t, conn), a).StandIn
	if si == nil || si.Tier != "heuristic" || si.Since != since.UTC().Format(time.RFC3339) {
		t.Fatalf("stand-in mark %+v, want heuristic since %v", si, since)
	}

	id := sendActionFrame(t, conn, protocol.ActionPayload{Type: "pass_priority", Player: a.String()})
	f := readFrame(t, conn)
	if f.Kind != protocol.KindError || f.ID != id {
		t.Fatalf("frame %s %q, want the error for %q", f.Kind, f.ID, id)
	}
	var e protocol.ErrorPayload
	if err := json.Unmarshal(f.Payload, &e); err != nil {
		t.Fatal(err)
	}
	if e.Code != protocol.CodeStandIn {
		t.Errorf("error code %q, want %q", e.Code, protocol.CodeStandIn)
	}

	// Taken back: the seat plays again, and the mark is gone.
	room.ClearStandIn(a)
	view := room.ViewFor(a.String())
	for _, s := range view.Seats {
		if s.StandIn != nil {
			t.Errorf("seat %s still marked after ClearStandIn", s.ID)
		}
	}
}

func TestSetMovableReportsAChange(t *testing.T) {
	_, g, room, _, _ := presenceServer(t)
	a := g.Seats[0].ID
	if !room.SetMovable(map[uuid.UUID]bool{a: true}) {
		t.Error("first mark: no change reported")
	}
	if room.SetMovable(map[uuid.UUID]bool{a: true}) {
		t.Error("the same mark again reported a change")
	}
	if !room.SetMovable(nil) || room.Movable(a) {
		t.Error("clearing the mark: no change, or still movable")
	}
}
