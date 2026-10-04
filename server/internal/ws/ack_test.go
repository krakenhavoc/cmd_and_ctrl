package ws

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ack_test.go — ADR 0122 §6.4. Every action the server applies is
// acknowledged to the connection that sent it, and only to it, on all
// three apply paths: an ordinary action (Room.Apply), a table-settings
// change (Room.ApplyExternal) and an undo (Room.Undo). The ack always
// arrives after the snapshot of the state it names. A refused action
// gets its error and never an ack.

// expectSilence fails if any frame arrives on conn within d. A read
// that times out leaves a gorilla connection unusable, so it is the
// last read a test makes on that connection.
func expectSilence(t *testing.T, conn *websocket.Conn, d time.Duration) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(d))
	_, raw, err := conn.ReadMessage()
	if err == nil {
		var f protocol.Frame
		_ = json.Unmarshal(raw, &f)
		t.Fatalf("want nothing more on this connection, got %q id=%q", f.Kind, f.ID)
	}
}

// ackAfterSnapshot reads the next two frames on conn and fails unless
// they are a snapshot and then the ack for id naming that snapshot.
func ackAfterSnapshot(t *testing.T, conn *websocket.Conn, id, what string) protocol.SnapshotPayload {
	t.Helper()
	f := readRawFrame(t, conn)
	if f.Kind != protocol.KindSnapshot {
		t.Fatalf("%s: first frame %q, want the snapshot (%s)", what, f.Kind, f.Payload)
	}
	var snap protocol.SnapshotPayload
	if err := json.Unmarshal(f.Payload, &snap); err != nil {
		t.Fatal(err)
	}
	ack := readRawFrame(t, conn)
	if ack.Kind != protocol.KindAck || ack.ID != id {
		t.Fatalf("%s: after the snapshot got %q id=%q, want ack id=%q", what, ack.Kind, ack.ID, id)
	}
	var p protocol.AckPayload
	if err := json.Unmarshal(ack.Payload, &p); err != nil {
		t.Fatal(err)
	}
	if p.Seq != snap.Seq || p.Generation != snap.Generation {
		t.Fatalf("%s: ack names seq %d gen %d, its snapshot is seq %d gen %d", what, p.Seq, p.Generation, snap.Seq, snap.Generation)
	}
	return snap
}

func TestAckFollowsItsSnapshotOnTheSenderOnly(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()
	host, other := g.Seats[0].ID, g.Seats[1].ID
	room.SetHost(host)

	connA := dialAs(t, wsURL, host)
	defer connA.Close()
	connB := dialAs(t, wsURL, other)
	defer connB.Close()
	readRawFrame(t, connA) // initial snapshots
	readRawFrame(t, connB)

	steps := []struct {
		what   string
		action protocol.ActionPayload
	}{
		{"an action (Room.Apply)", protocol.ActionPayload{Type: "draw_card", Player: host.String()}},
		{"a settings change (Room.ApplyExternal)", settingsAction(`{"undo_limit":5}`)},
		{"an undo (Room.Undo)", protocol.ActionPayload{Type: "undo"}},
	}
	var seqs []uint64
	for _, s := range steps {
		id := sendActionFrame(t, connA, s.action)
		snap := ackAfterSnapshot(t, connA, id, s.what)
		seqs = append(seqs, snap.Seq)
	}

	// The other seat saw each state's snapshot and no ack.
	for i, s := range steps {
		f := readRawFrame(t, connB)
		if f.Kind != protocol.KindSnapshot {
			t.Fatalf("other seat, after %s: got %q, want only the snapshot", s.what, f.Kind)
		}
		var snap protocol.SnapshotPayload
		if err := json.Unmarshal(f.Payload, &snap); err != nil {
			t.Fatal(err)
		}
		if snap.Seq != seqs[i] {
			t.Fatalf("other seat's snapshot %d is seq %d, want %d", i, snap.Seq, seqs[i])
		}
	}

	// Refused on each path: an unknown action, an undo with nothing to
	// undo, and a settings change from a seat that does not host. Each
	// gets its error with the action's id, and no ack.
	for _, r := range []struct {
		conn   *websocket.Conn
		action protocol.ActionPayload
	}{
		{connA, protocol.ActionPayload{Type: "no_such_action"}},
		{connA, protocol.ActionPayload{Type: "undo"}},
		{connB, settingsAction(`{"undo_limit":2}`)},
	} {
		id := sendActionFrame(t, r.conn, r.action)
		f := readRawFrame(t, r.conn)
		if f.Kind != protocol.KindError || f.ID != id {
			t.Fatalf("refused %q: got %q id=%q, want its error", r.action.Type, f.Kind, f.ID)
		}
	}
	expectSilence(t, connA, 150*time.Millisecond)
	expectSilence(t, connB, 150*time.Millisecond)
}
