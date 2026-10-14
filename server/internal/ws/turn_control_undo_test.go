package ws

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0167 §4: an action the controller sends as the player whose turn
// they play is stamped with the controller's seat, the seat of the
// connection that sent it. It is theirs to undo, from their own budget,
// and the controlled player can't rewind it.
func TestTurnControlUndoBelongsToTheController(t *testing.T) {
	g := game.NewGame()
	for i := range 3 {
		deck := []game.Card{game.NewCommander(fmt.Sprintf("Cmdr %d", i+1), uuid.Nil)}
		for j := range 20 {
			deck = append(deck, game.NewCard(fmt.Sprintf("Filler %d", j+1), uuid.Nil))
		}
		if _, err := g.AddPlayer(fmt.Sprintf("P%d", i+1), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(5, 6))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	a := g.Turn.ActiveSeat
	alice, bob := g.Seats[a].ID, g.Seats[(a+1)%3].ID
	g.WithWriteLock(func() {
		g.Settings.UndoLimit = game.UndoUnlimited
		g.BeginTurnControlForEffect(game.TurnControl{Controller: alice, Controlled: bob, Source: uuid.New(), SourceName: "Emrakul, the Promised End"})
	})
	if err := g.PassTurn(); err != nil {
		t.Fatalf("PassTurn: %v", err)
	}
	room := NewRoom(g, quietLog(), "")

	// What the hub does with a frame from Alice's connection: the tag's
	// Actor is the connection's seat.
	pass := actions.Action{Type: actions.TypePassPriority, Player: bob, Caller: alice}
	if _, _, err := room.ApplyAction(ActionTag{Type: pass.Type, Actor: alice}, func() error {
		return actions.Dispatch(room.Game, pass)
	}); err != nil {
		t.Fatalf("Alice passing for Bob: %v", err)
	}
	holder := func() uuid.UUID { return g.Seats[g.Turn.PriorityHolder].ID }
	if holder() == bob {
		t.Fatal("the pass did not move priority")
	}
	if _, _, err := room.Undo(bob); !errors.Is(err, ErrNotYourUndo) {
		t.Fatalf("Bob undoing Alice's play as him: %v, want ErrNotYourUndo", err)
	}
	if _, _, err := room.Undo(alice); err != nil {
		t.Fatalf("Alice undoing her play as Bob: %v", err)
	}
	if holder() != bob {
		t.Error("the undo did not give Bob priority back")
	}
	if g.ControllerOfPlayer(bob) != alice {
		t.Error("the undo lost the control record")
	}
}
