package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// generation_wire_test.go covers the OTHER half of #523:
// persist_test.go proves Room.Generation() is right in memory; this
// proves the value actually reaches the wire, on both the frame a
// just-connected client is staged and every later broadcast — the two
// paths marshalSnapshotFrame is shared between.
//
// TestSnapshotFrameCarriesGeneration mutates room.generation directly
// rather than restoring a room from disk. That is deliberate: this
// file's job is the room → wire leg, which is exactly the same
// regardless of how the room came to hold that generation, and
// persist_test.go already covers restoreOne's half.
func TestSnapshotFrameCarriesGeneration(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()

	// Nothing has dialled in yet, so this direct write is race-free.
	room.generation = 3

	seat0 := g.Seats[0].ID
	conn := dialAs(t, wsURL, seat0)
	defer func() { _ = conn.Close() }()

	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read initial snapshot: %v", err)
	}
	var f protocol.Frame
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("unmarshal frame: %v", err)
	}
	if f.Kind != protocol.KindSnapshot {
		t.Fatalf("kind = %q, want snapshot", f.Kind)
	}
	var snap protocol.SnapshotPayload
	if err := json.Unmarshal(f.Payload, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if snap.Generation != 3 {
		t.Errorf("initial snapshot generation = %d, want 3", snap.Generation)
	}

	// A broadcast after an ordinary action carries the same
	// generation — it is a room-lifetime constant, not something
	// that varies action to action.
	snap = sendActionAndWait(t, conn, protocol.ActionPayload{Type: "draw_card", Player: seat0.String()})
	if snap.Generation != 3 {
		t.Errorf("post-action snapshot generation = %d, want 3", snap.Generation)
	}
}
