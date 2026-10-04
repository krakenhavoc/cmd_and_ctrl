package ws

// table_roll_test.go — ADR 0121 §5 at the WebSocket edge: a table roll
// mints no undo entry, a seat rolls at most once per 2 s, a refused
// roll does not spend that slot, and an undo of another seat's earlier
// action keeps the roll's line.

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func tableRollLog(log []protocol.LogEvent) []protocol.LogEvent {
	var out []protocol.LogEvent
	for _, e := range log {
		if e.Kind == protocol.LogTableRoll {
			out = append(out, e)
		}
	}
	return out
}

func TestTableRollOverTheWire(t *testing.T) {
	wsURL, g, room, cleanup := newE2EServerWithRoom(t)
	defer cleanup()
	// The limiter's clock, moved by hand. Set before any connection, so
	// the hub's goroutines only ever read it.
	var elapsed atomic.Int64
	base := time.Unix(1_000_000, 0)
	room.tableRolls.now = func() time.Time { return base.Add(time.Duration(elapsed.Load())) }

	a := dialAs(t, wsURL, g.Seats[0].ID)
	defer a.Close()
	readNextFrame(t, a)
	b := dialAs(t, wsURL, g.Seats[1].ID)
	defer b.Close()
	readNextFrame(t, b)
	roll := func(die string) protocol.ActionPayload {
		return protocol.ActionPayload{
			Type: "roll_table_die", Player: g.Seats[1].ID.String(),
			Params: json.RawMessage(`{"die":"` + die + `"}`),
		}
	}

	// A's action, which A may undo.
	sendActionAndWait(t, a, protocol.ActionPayload{Type: "keep_hand", Player: g.Seats[0].ID.String()})
	readNextFrame(t, b)
	if n := undoDepth(room); n != 1 {
		t.Fatalf("keep_hand left %d undo entries, want 1", n)
	}

	// A refused roll does not spend the seat's slot.
	ep := expectActionError(t, b, roll("d100"))
	if ep.Code != protocol.CodeBadRequest || !strings.Contains(ep.Message, "d6, a d20 or a coin") {
		t.Fatalf("a d100: %+v", ep)
	}
	snap := sendActionAndWait(t, b, roll("d20"))
	readNextFrame(t, a)
	lines := tableRollLog(snap.Game.Log)
	if len(lines) != 1 || lines[0].RollID != 1 || lines[0].Sides != 20 ||
		!strings.HasPrefix(lines[0].Text, "Player 2 rolled a d20 at the table: ") {
		t.Fatalf("after the roll: %+v", lines)
	}
	if n := undoDepth(room); n != 1 {
		t.Fatalf("a table roll minted an undo entry: %d, want 1", n)
	}

	// The next one inside 2 s is refused.
	elapsed.Store(int64(1999 * time.Millisecond))
	ep = expectActionError(t, b, roll("coin"))
	if ep.Code != protocol.CodeBadRequest || ep.Message != "wait for your last roll to land" {
		t.Fatalf("a second roll inside 2 s: %+v", ep)
	}

	// A undoes their own earlier action. B's roll, made after it, keeps
	// its line and its roll_id.
	snap = sendActionAndWait(t, a, protocol.ActionPayload{Type: "undo"})
	readNextFrame(t, b)
	after := tableRollLog(snap.Game.Log)
	if len(after) != 1 || after[0].RollID != 1 || after[0].Text != lines[0].Text {
		t.Fatalf("after A's undo: %+v, want B's line %q", after, lines[0].Text)
	}
	if snap.Game.Seats[0].HandKept {
		t.Fatal("the undo did not take back A's keep")
	}

	// 2 s on, B rolls again: a fresh roll.
	elapsed.Store(int64(2 * time.Second))
	snap = sendActionAndWait(t, b, roll("coin"))
	if got := tableRollLog(snap.Game.Log); len(got) != 2 || got[1].RollID != 2 || len(got[1].Faces) != 1 {
		t.Fatalf("after the second roll: %+v", got)
	}
}
