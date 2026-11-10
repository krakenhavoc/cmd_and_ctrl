package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ironShieldElfOracle = "1c77b4da-b580-4ce9-b127-76de674f5694"

func TestIronShieldElfIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(ironShieldElfOracle)
	if !ok || spec.Name != "Iron-Shield Elf" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// Discarding a card gives the Elf indestructible and taps it.
func TestIronShieldElfDiscardsForIndestructibleAndTaps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	elf := b12Push(g, me.ID, "Iron-Shield Elf", "Creature — Elf Warrior", ironShieldElfOracle, 3, 1)
	ids := fillHandForTest(g, me, 2)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{ids[0]}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Size() != 1 {
		t.Errorf("hand %d after the discard cost, want 1", me.Hand.Size())
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, elf, "indestructible") {
		t.Error("the Elf did not gain indestructible")
	}
	if !tappedNow(g, elf) {
		t.Error("the Elf did not tap")
	}
}

// The ruling: a tapped Elf can still activate it. With no card in hand
// the cost can't be paid.
func TestIronShieldElfActivatesWhileTappedButNeedsACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMainForCost(t, g)
	elf := b12Push(g, me.ID, "Iron-Shield Elf", "Creature — Elf Warrior", ironShieldElfOracle, 3, 1)
	tapForTest(t, g, elf)
	emptyHandForTest(g, me)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with an empty hand")
	}
	ids := fillHandForTest(g, me, 1)
	if err := g.ActivateCatalogAbility(me.ID, elf, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{ids[0]}}); err != nil {
		t.Fatalf("activate while tapped: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, elf, "indestructible") {
		t.Error("the tapped Elf did not gain indestructible")
	}
}
