package ws

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// game_over_undo_test.go — ADR 0057 Decision 5, test plan item 16, the
// owner's answer to question 1: the end of a game is final. After a
// concession or an effect win ends the game, a player's undo is
// refused and the game stays ended; the admin's undo still works.

func endedRoom(t *testing.T, end func(g *game.Game, caller uuid.UUID) error) (*Room, *game.Game, uuid.UUID) {
	t.Helper()
	g := seedTestGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatal(err)
		}
	}
	room := NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	caller := g.Seats[0].ID
	if _, _, err := room.Apply(caller, func() error { return end(g, caller) }); err != nil {
		t.Fatalf("ending action: %v", err)
	}
	if g.CurrentState() != game.StateEnded {
		t.Fatalf("the action did not end the game")
	}
	return room, g, caller
}

func TestUndoIsRefusedOnceTheGameHasEnded(t *testing.T) {
	cases := map[string]func(g *game.Game, caller uuid.UUID) error{
		"concede": func(g *game.Game, caller uuid.UUID) error { return g.Concede(caller) },
		"effect win": func(g *game.Game, caller uuid.UUID) error {
			var err error
			g.WithWriteLock(func() { _, err = g.WinTheGameForEffect(caller, uuid.Nil) })
			return err
		},
	}
	for name, end := range cases {
		t.Run(name, func(t *testing.T) {
			room, g, caller := endedRoom(t, end)
			if _, _, err := room.Undo(caller); !errors.Is(err, ErrGameOverUndo) {
				t.Fatalf("player undo after the end: %v, want ErrGameOverUndo", err)
			}
			if g.CurrentState() != game.StateEnded {
				t.Fatalf("the refused undo reopened the game")
			}
			// The admin can still undo, for mistakes.
			if _, _, err := room.Undo(uuid.Nil); err != nil {
				t.Fatalf("admin undo: %v", err)
			}
			if g.CurrentState() != game.StateActive || g.Result() != nil {
				t.Fatalf("admin undo left state %s outcome %+v", g.CurrentState(), g.Result())
			}
		})
	}
}
