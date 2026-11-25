package testss82m

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB163GleanerOracle = "e26f7255-8b18-47d6-aafa-22ca22c35d2e"

func TestOldB163FarmsteadGleanerUntapsToGrowAndOtherwiseStaysTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gleaner := effectstest.PushCatalogPermanent(g, me.ID, "Farmstead Gleaner", "Artifact Creature — Scarecrow", oldB163GleanerOracle, false)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gleaner, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated {Q} while untapped")
	}
	effectstest.AvB4SetTapped(g, gleaner)
	if err := g.ActivateCatalogAbility(me.ID, gleaner, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, gleaner, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if effectstest.B16Tapped(t, g, gleaner) {
		t.Error("the Gleaner stayed tapped: {Q} untaps it")
	}
}

func TestOldB163FarmsteadGleanerDoesNotUntapDuringYourUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gleaner := effectstest.PushCatalogPermanent(g, me.ID, "Farmstead Gleaner", "Artifact Creature — Scarecrow", oldB163GleanerOracle, false)
	effectstest.AvB4SetTapped(g, gleaner)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.B16Tapped(t, g, gleaner) {
		t.Error("the Gleaner untapped during its controller's untap step")
	}
}
