package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// blb_forage_cards_test.go — #5184: the forage keyword action (CR
// 701.61a) as an activated ability's cost (Camellia, the Seedmiser), as
// an either/or additional cost (Feed the Cycle), as an effect's "you may
// forage" (Treetop Sentries, Bushy Bodyguard, Curious Forager, Corpseberry
// Cultivator),
// and "whenever you forage" (Corpseberry Cultivator). Bloomburrow, #4467.
// Every test drives the real catalog card.

const (
	BlbFgCamelliaOracle    = "312c7a7b-2084-42b0-9b13-0a6d3e43a73d"
	BlbFgCorpseberryOracle = "5ded56df-c7d5-4043-85df-410f47219faa"
	BlbFgCuriousOracle     = "30af656a-99ab-4e0a-81fe-3cdb0ea6a0cc"
	BlbFgFeedOracle        = "c338eedb-9b70-4b77-b5d6-8ea74b7bea3e"
	BlbFgTreetopOracle     = "5bf5739a-6322-4e6e-b41e-e7bdbb11b1ce"
	BlbFgBushyOracle       = "5436bb0d-b5c7-43a9-8b3f-67cdc0dd0381"
)

// fgGraveyard puts n cards in p's graveyard and returns them, oldest
// first.
func FgGraveyard(p *game.Player, n int) []uuid.UUID {
	out := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		c := game.NewCard("Old Card", p.ID)
		c.TypeLine = "Instant"
		p.Graveyard.PushTop(c)
		out = append(out, c.InstanceID)
	}
	return out
}

// fgFoodToken creates a Food token under owner and returns it.
func FgFoodToken(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, effects.FoodToken(), 1); err != nil {
			t.Fatal(err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] && c.Name == "Food" {
			return c.InstanceID
		}
	}
	t.Fatal("no Food token was created")
	return uuid.Nil
}

// fgForages counts the EventForages player has made.
func FgForages(g *game.Game, player uuid.UUID) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventForage && ev.Actor == player {
			n++
		}
	}
	return n
}

func FgPlusOnes(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Counters[game.CounterPlusOne]
}
