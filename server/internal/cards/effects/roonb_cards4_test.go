package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// roonb_cards4_test.go — #4128 slice b: Blatant Thievery at a
// two-player table, where only the first per-opponent clause applies.

func roonbTwoPlayerGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func TestRoonbBlatantThieveryAtATwoPlayerTableNeedsOnePick(t *testing.T) {
	g := roonbTwoPlayerGame(t)
	rock := roonbPermanent(g, g.Seats[1].ID, "Rock")
	if err := roonbThieveryCast(t, g, rock); err != nil {
		t.Fatalf("one opponent, one target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if roonbControllerOf(g, rock) != g.Seats[0].ID {
		t.Error("the opponent's permanent was not taken")
	}
}

func TestRoonbBlatantThieveryAtATwoPlayerTableRefusesASecondPick(t *testing.T) {
	g := roonbTwoPlayerGame(t)
	rock := roonbPermanent(g, g.Seats[1].ID, "Rock")
	mine := roonbPermanent(g, g.Seats[0].ID, "Mine")
	if err := roonbThieveryCast(t, g, rock, mine); err == nil {
		t.Fatal("a second pick has no opponent to belong to and must be refused")
	}
}
