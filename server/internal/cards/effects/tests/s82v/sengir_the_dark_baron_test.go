package testss82v

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sengir_the_dark_baron_test.go — Sengir, the Dark Baron (#6164).

const sengirOracle = "f49f84f9-616e-495b-8d87-4de7941f230e"

func TestSengirIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(sengirOracle)
	if !ok || spec.Name != "Sengir, the Dark Baron" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// An opponent at 12 as the turn began who loses at −3 gives Sengir's
// controller 12 life.
func TestSengirGainsTheLosersLifeAsTheTurnBegan(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	next := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	opp := g.Seats[next]
	effectstest.B27Push(g, me.ID, "Sengir, the Dark Baron", "Legendary Creature — Vampire Noble", sengirOracle, "{4}{B}{B}", 4, 4, "B")
	g.WithWriteLock(func() { opp.Life = 12 })
	effectstest.AdvanceToMainOf(t, g, next)
	g.WithWriteLock(func() { opp.Life = -3 })
	before := effectstest.LifeOf(g, me.ID)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Eliminated {
		t.Fatal("the opponent at -3 did not lose the game")
	}
	if got := effectstest.LifeOf(g, me.ID) - before; got != 12 {
		t.Errorf("Sengir's controller gained %d life, want 12", got)
	}
}

// Sengir's controller losing gains nobody anything from this Sengir.
func TestSengirIgnoresItsControllersOwnLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sengir := effectstest.B27Push(g, me.ID, "Sengir, the Dark Baron", "Legendary Creature — Vampire Noble", sengirOracle, "{4}{B}{B}", 4, 4, "B")
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	before := effectstest.LifeOf(g, other.ID)
	g.WithWriteLock(func() { me.Life = 0 })
	g.RunStateChecksForTest()
	if !me.Eliminated {
		t.Fatal("Sengir's controller at 0 did not lose")
	}
	if got := effectstest.LifeOf(g, other.ID); got != before {
		t.Errorf("another player's life moved from %d to %d", before, got)
	}
	_ = sengir
}

func TestSengirGrowsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sengir := effectstest.B27Push(g, me.ID, "Sengir, the Dark Baron", "Legendary Creature — Vampire Noble", sengirOracle, "{4}{B}{B}", 4, 4, "B")
	bear := effectstest.B27Push(g, opp.ID, "Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear, game.DestroyOptions{}) })
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, sengir); p != 6 || tough != 6 {
		t.Errorf("Sengir is %d/%d after a creature died, want 6/6", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, sengir, "flying") {
		t.Error("Sengir does not have flying")
	}
}
