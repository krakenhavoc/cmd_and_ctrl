package aiseat

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// opening_roll_internal_test.go — ADR 0121 §3: the bot runner is the
// second caller of Room.Apply, and routes what actions.MintsNoUndo names
// through Room.ApplyExternal exactly as the hub does.
func TestRunnerAppliesTheOpeningRollWithoutAnUndoEntry(t *testing.T) {
	g := game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 10 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("Bot %d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithOpeningRoll(rand.New(rand.NewPCG(1, 2027))); err != nil { // 20 17 1
		t.Fatal(err)
	}
	room := ws.NewRoom(g, slog.New(slog.NewTextHandler(io.Discard, nil)), "")
	bot := func(seat int) *Runner { return &Runner{room: room, seat: g.Seats[seat].ID} }

	for seat := range 3 {
		if _, _, err := bot(seat).apply(actions.Action{Type: actions.TypeRollOpening, Player: g.Seats[seat].ID, Caller: g.Seats[seat].ID}); err != nil {
			t.Fatalf("seat %d roll_opening: %v", seat, err)
		}
	}
	if _, _, err := bot(2).apply(actions.Action{Type: actions.TypeConcede, Player: g.Seats[2].ID, Caller: g.Seats[2].ID}); err != nil {
		t.Fatalf("concede during the roll: %v", err)
	}
	if _, _, err := room.Undo(uuid.Nil); !errors.Is(err, ws.ErrNothingToUndo) {
		t.Fatalf("undo after the roll and a concession: %v, want ErrNothingToUndo", err)
	}
	if _, _, err := bot(0).apply(actions.Action{
		Type: actions.TypeChooseStartingPlayer, Player: g.Seats[0].ID, Caller: g.Seats[0].ID, Params: []byte(`{"seat":0}`),
	}); err != nil {
		t.Fatalf("choose_starting_player: %v", err)
	}
	if _, _, err := room.Undo(uuid.Nil); !errors.Is(err, ws.ErrNothingToUndo) {
		t.Fatalf("undo after the choice: %v, want ErrNothingToUndo", err)
	}
	// An ordinary action after the deal is undoable again.
	if _, _, err := bot(1).apply(actions.Action{Type: actions.TypeKeepHand, Player: g.Seats[1].ID, Caller: g.Seats[1].ID}); err != nil {
		t.Fatalf("keep_hand: %v", err)
	}
	if _, _, err := room.Undo(uuid.Nil); err != nil {
		t.Fatalf("undo of keep_hand: %v", err)
	}
}
