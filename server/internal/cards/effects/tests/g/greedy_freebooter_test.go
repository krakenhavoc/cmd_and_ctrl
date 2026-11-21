package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const greedyFreebooterOracle = "d4a9d63a-5a8a-4f51-9e49-43d6dc3234b1"

// TestGreedyFreebooterDiesScriesAndMakesTreasure checks the dies
// trigger's scry and Treasure token.
func TestGreedyFreebooterDiesScriesAndMakesTreasure(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Bottom Me")
	freebooter := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Greedy Freebooter", TypeLine: "Creature — Human Pirate",
		OracleID: greedyFreebooterOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})

	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(freebooter) })
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no scry-1 prompt from the dies trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolveScry(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}

	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" {
			found = true
		}
	}
	if !found {
		t.Error("Greedy Freebooter should leave behind a Treasure token when it dies")
	}
}
