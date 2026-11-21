package effectstest

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// newTestGame returns a fully-started 2-player game suitable for
// primitive unit tests. Uses the game package's exported Start
// path via a tiny deck so we don't have to reimplement setup.
func NewTestGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	// Two players with 10-card libraries — enough for draw / mill
	// primitives without running out.
	for i := 0; i < 2; i++ {
		deck := make([]game.Card, 10)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		p, err := g.AddPlayer("P"+string(rune('0'+i)), deck)
		if err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
		_ = p
	}
	if err := g.Start(rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	// Close mulligans so the cursor lands on a priority-granting step.
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

func PushBattlefieldCreature(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Creature",
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

func PushHandCard(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Card",
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

func CtxFor(g *game.Game, item *game.StackItem) *effects.Context {
	return effects.NewContext(g, item)
}
