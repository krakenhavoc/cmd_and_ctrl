package ws

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// undo_barrier_test.go — ADR 0194: a bot's commit leaves a barrier
// instead of a clone, nobody (admin included) can undo past it, and
// only the top undoSnapshotDepth entries keep their clone.

// barrierRoom is a started two-seat room whose second seat is a bot.
func barrierRoom(t *testing.T) (r *Room, g *game.Game, human, bot uuid.UUID) {
	t.Helper()
	g = seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	g.WithWriteLock(func() { g.Seats[1].IsBot = true })
	r = NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	return r, g, g.Seats[0].ID, g.Seats[1].ID
}

// bumpLife is a commit with a visible effect: seat i's life goes up by one.
func bumpLife(g *game.Game, i int) func() error {
	return func() error {
		g.WithWriteLock(func() { g.Seats[i].Life++ })
		return nil
	}
}

func lifeOf(g *game.Game, i int) int {
	var n int
	g.WithWriteLock(func() { n = g.Seats[i].Life })
	return n
}

// clonesHeld counts the ring's entries that hold a full snapshot.
func clonesHeld(r *Room) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, e := range r.undoStack {
		if e.pre != nil {
			n++
		}
	}
	return n
}

func TestUndoPastABotIsRefusedForEveryone(t *testing.T) {
	room, g, human, bot := barrierRoom(t)

	if _, _, err := room.Apply(human, bumpLife(g, 0)); err != nil {
		t.Fatal(err)
	}
	v, _, err := room.Apply(bot, bumpLife(g, 1))
	if err != nil {
		t.Fatal(err)
	}
	if got := clonesHeld(room); got != 1 {
		t.Fatalf("the ring holds %d clones, want 1: the bot's commit takes none", got)
	}
	if got := blockedFor(t, v, human); got != protocol.UndoBlockedPastBot {
		t.Fatalf("human's stamp after a bot move: %q, want %q", got, protocol.UndoBlockedPastBot)
	}
	if got := blockedFor(t, v, bot); got != protocol.UndoBlockedPastBot {
		t.Fatalf("bot's stamp after its own move: %q, want %q", got, protocol.UndoBlockedPastBot)
	}

	humanLife, botLife := lifeOf(g, 0), lifeOf(g, 1)
	for _, who := range []struct {
		name   string
		caller uuid.UUID
	}{{"the human", human}, {"the bot", bot}, {"the admin", uuid.Nil}} {
		_, _, err := room.Undo(who.caller)
		if !errors.Is(err, ErrUndoPastBot) {
			t.Fatalf("undo by %s past a bot move: %v, want ErrUndoPastBot", who.name, err)
		}
	}
	if err := ErrUndoPastBot; err.Error() != "ws: a bot has acted since; a bot's move can't be taken back" {
		t.Fatalf("message %q", err)
	}
	if lifeOf(g, 0) != humanLife || lifeOf(g, 1) != botLife {
		t.Fatal("a refused undo changed the game")
	}

	// The human acts again: their own move is on top and undoable, and
	// the barrier is back on top after it.
	if _, _, err := room.Apply(human, bumpLife(g, 0)); err != nil {
		t.Fatal(err)
	}
	if _, _, err := room.Undo(human); err != nil {
		t.Fatalf("undo of the human's move over a barrier: %v", err)
	}
	if lifeOf(g, 0) != humanLife {
		t.Fatalf("life %d after the undo, want %d", lifeOf(g, 0), humanLife)
	}
	if _, _, err := room.Undo(uuid.Nil); !errors.Is(err, ErrUndoPastBot) {
		t.Fatalf("admin undo of the barrier: %v, want ErrUndoPastBot", err)
	}
}

func TestBotBundleLeavesABarrierAndAnImprovisationDoesNot(t *testing.T) {
	room, g, human, bot := barrierRoom(t)

	// A failing step under a bot's stamp still rolls back to the clone.
	before := lifeOf(g, 1)
	_, _, err := room.ApplyBundle(Bundle{Caller: bot, Actor: bot, Steps: []func() error{
		bumpLife(g, 1),
		func() error { return errors.New("boom") },
	}})
	if err == nil {
		t.Fatal("the failing bundle committed")
	}
	if lifeOf(g, 1) != before {
		t.Fatal("a failed bot bundle was not rolled back")
	}

	if _, _, err := room.ApplyBundle(Bundle{Caller: bot, Actor: bot, Steps: []func() error{bumpLife(g, 1)}}); err != nil {
		t.Fatal(err)
	}
	if got := clonesHeld(room); got != 0 {
		t.Fatalf("a bot-stamped bundle kept %d clones, want 0", got)
	}
	if _, _, err := room.Undo(uuid.Nil); !errors.Is(err, ErrUndoPastBot) {
		t.Fatalf("admin undo of a bot bundle: %v, want ErrUndoPastBot", err)
	}

	// An improvisation is nil-stamped (ADR 0033 §8): any seat can take
	// it back, so it keeps its clone.
	before = lifeOf(g, 1)
	if _, _, err := room.ApplyBundle(Bundle{Caller: uuid.Nil, Actor: bot, FreeUndo: true, Steps: []func() error{bumpLife(g, 1)}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := room.Undo(human); err != nil {
		t.Fatalf("human undo of an improvisation: %v", err)
	}
	if lifeOf(g, 1) != before {
		t.Fatal("the improvisation was not taken back")
	}
}

func TestUndoGoesBackEightActions(t *testing.T) {
	room, g, human, _ := barrierRoom(t)
	start := lifeOf(g, 0)
	const commits = 12
	for i := 0; i < commits; i++ {
		if _, _, err := room.Apply(human, bumpLife(g, 0)); err != nil {
			t.Fatal(err)
		}
		if want := min(i+1, undoSnapshotDepth); clonesHeld(room) != want {
			t.Fatalf("after %d commits the ring holds %d clones, want %d", i+1, clonesHeld(room), want)
		}
	}
	if undoSnapshotDepth != 8 {
		t.Fatalf("undoSnapshotDepth = %d, want 8 (ADR 0194)", undoSnapshotDepth)
	}

	// The admin (no budget) takes back eight, one at a time.
	for i := 1; i <= undoSnapshotDepth; i++ {
		if _, _, err := room.Undo(uuid.Nil); err != nil {
			t.Fatalf("undo %d of %d: %v", i, undoSnapshotDepth, err)
		}
		if want := start + commits - i; lifeOf(g, 0) != want {
			t.Fatalf("after undo %d life is %d, want %d", i, lifeOf(g, 0), want)
		}
	}

	// The ninth is past the depth, for the admin and the owner alike.
	v, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if got := blockedFor(t, v, human); got != protocol.UndoBlockedTooDeep {
		t.Fatalf("stamp past the depth: %q, want %q", got, protocol.UndoBlockedTooDeep)
	}
	for _, caller := range []uuid.UUID{uuid.Nil, human} {
		if _, _, err := room.Undo(caller); !errors.Is(err, ErrUndoTooDeep) {
			t.Fatalf("ninth undo by %v: %v, want ErrUndoTooDeep", caller, err)
		}
	}
	if want := start + commits - undoSnapshotDepth; lifeOf(g, 0) != want {
		t.Fatalf("a refused undo changed life to %d, want %d", lifeOf(g, 0), want)
	}
}

// TestBotTableRingHoldsNoClones is the memory claim: at a table of bots
// the ring holds nothing but barriers, however long the game runs.
func TestBotTableRingHoldsNoClones(t *testing.T) {
	room, g, human, bot := barrierRoom(t)
	g.WithWriteLock(func() { g.Seats[0].IsBot = true })
	for i := 0; i < 2*undoStackCap; i++ {
		seat := []uuid.UUID{human, bot}[i%2]
		if _, _, err := room.Apply(seat, bumpLife(g, i%2)); err != nil {
			t.Fatal(err)
		}
	}
	room.mu.Lock()
	n := len(room.undoStack)
	room.mu.Unlock()
	if n != undoStackCap {
		t.Fatalf("ring length %d, want %d", n, undoStackCap)
	}
	if got := clonesHeld(room); got != 0 {
		t.Fatalf("a table of bots holds %d clones, want 0", got)
	}
}
