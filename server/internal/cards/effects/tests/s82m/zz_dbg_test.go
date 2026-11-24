package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestZZDbgChandra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	chandra := effectstest.S82Walker(g, me.ID, "Chandra, Fire Artisan", s82OldB5ChandraOracle, 4)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(chandra, game.CounterLoyalty, -2) })
	g.RunStateChecksForTest()
	t.Logf("stack %d pending %d life %d", g.Stack.Size(), len(g.PendingChoices), opp.Life)
	for _, c := range g.PendingChoices {
		t.Logf("choice %v chooser %v", c.Kind, c.Chooser)
	}
}

func TestZZDbgVraska(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vraska := effectstest.S82Walker(g, me.ID, "Vraska, Golgari Queen", s82OldB5VraskaOracle, 4)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, vraska, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		t.Logf("choice %v chooser %v opts %d", c.Kind, c.Chooser, len(c.ChooseCards))
	}
}
