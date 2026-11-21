package testsi

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ironShieldElfOracle = "1c77b4da-b580-4ce9-b127-76de674f5694"

func TestIronShieldElfIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(ironShieldElfOracle)
	if !ok || spec.Name != "Iron-Shield Elf" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Discarding a card gives the Elf indestructible and taps it.
func TestIronShieldElfDiscardsForIndestructibleAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	elf := effectstest.B12Push(g, me.ID, "Iron-Shield Elf", "Creature — Elf Warrior", ironShieldElfOracle, 3, 1)
	ids := effectstest.FillHandForTest(g, me, 2)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{ids[0]}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d after the discard cost, want 1", me.Hand.Size())
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, elf, "indestructible") {
		t.Error("the Elf did not gain indestructible")
	}
	if !effectstest.TappedNow(g, elf) {
		t.Error("the Elf did not tap")
	}
}

// The ruling: a tapped Elf can still activate it. With no card in hand
// the cost can't be paid.
func TestIronShieldElfActivatesWhileTappedButNeedsACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	elf := effectstest.B12Push(g, me.ID, "Iron-Shield Elf", "Creature — Elf Warrior", ironShieldElfOracle, 3, 1)
	effectstest.TapForTest(t, g, elf)
	effectstest.EmptyHandForTest(g, me)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty hand")
	}
	ids := effectstest.FillHandForTest(g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{ids[0]}}); err != nil {
		t.Fatalf("activate while tapped: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, elf, "indestructible") {
		t.Error("the tapped Elf did not gain indestructible")
	}
}
