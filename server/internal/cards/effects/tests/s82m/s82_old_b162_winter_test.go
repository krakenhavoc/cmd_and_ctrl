package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB162DeadOfWinterOracle = "4d877c28-5276-4ea6-91c6-eeed890ca545"

func TestOldB162DeadOfWinterShrinksNonsnowCreaturesBySnowPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	// A snow creature is a snow permanent, counts, and is not shrunk.
	yeti := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Snow Yeti", TypeLine: "Snow Creature — Yeti", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	bear := oldB162Bear(g, me.ID, "Bear")
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Dead of Winter", "Sorcery", oldB162DeadOfWinterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the 2/2 Bear survived -3/-3")
	}
	if p, tough := effectstest.EffectivePower(t, g, big), effectstest.EffectiveToughness(t, g, big); p != 1 || tough != 1 {
		t.Errorf("their Ogre is %d/%d, want 1/1", p, tough)
	}
	if p := effectstest.EffectivePower(t, g, yeti); p != 3 {
		t.Errorf("the snow creature has power %d, want 3", p)
	}
}

func TestOldB162DeadOfWinterWithNoSnowDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := oldB162Bear(g, me.ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Dead of Winter", "Sorcery", oldB162DeadOfWinterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) || effectstest.EffectivePower(t, g, bear) != 2 {
		t.Error("the Bear changed with X = 0")
	}
}
