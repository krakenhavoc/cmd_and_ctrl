package ws

import (
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// undo_blocked_test.go — #2963: the view tells each seat why an undo
// would be refused, from the same gates room.undo applies, so the
// client never offers an undo the server turns away.

func blockedFor(t *testing.T, v protocol.GameView, id uuid.UUID) string {
	t.Helper()
	for _, s := range v.Seats {
		if s.ID == id.String() {
			return s.UndoBlocked
		}
	}
	t.Fatalf("seat %s not in view", id)
	return ""
}

func TestViewSaysWhyAnUndoWouldBeRefused(t *testing.T) {
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	a, b := g.Seats[0].ID, g.Seats[1].ID

	v, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	if got := blockedFor(t, v, a); got != protocol.UndoBlockedNothing {
		t.Fatalf("empty stack: %q, want nothing", got)
	}

	// A's action is on top: A may undo it, B may not.
	v, _, err = room.Apply(a, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if got := blockedFor(t, v, a); got != "" {
		t.Fatalf("owner: %q, want allowed", got)
	}
	if got := blockedFor(t, v, b); got != protocol.UndoBlockedNotYours {
		t.Fatalf("other seat: %q, want not_yours", got)
	}

	// An admin-origin entry (caller Nil) is undoable by anyone.
	v, _, err = room.Apply(uuid.Nil, func() error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if blockedFor(t, v, a) != "" || blockedFor(t, v, b) != "" {
		t.Fatal("admin entry should be undoable by every seat")
	}

	// The end is final for seats, whoever owns the top entry.
	v, _, err = room.Apply(a, func() error { return g.Concede(a) })
	if err != nil {
		t.Fatal(err)
	}
	if g.CurrentState() != game.StateEnded {
		t.Fatal("game did not end")
	}
	if got := blockedFor(t, v, a); got != protocol.UndoBlockedGameOver {
		t.Fatalf("after the end: %q, want game_over", got)
	}

	// The stamp agrees with the gate: what it blocks, Undo refuses.
	if _, _, err := room.Undo(a); err == nil {
		t.Fatal("undo after the end should be refused")
	}
}

func TestUndoBlockedIsPrivateToItsSeat(t *testing.T) {
	g := seedTestGame(t)
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	v, _, err := room.Snapshot()
	if err != nil {
		t.Fatal(err)
	}
	a, b := g.Seats[0].ID, g.Seats[1].ID
	f := protocol.FilterViewFor(v, a.String())
	if blockedFor(t, f, a) == "" {
		t.Fatal("own seat lost its reason")
	}
	if got := blockedFor(t, f, b); got != "" {
		t.Fatalf("opponent's reason leaked: %q", got)
	}
}
