package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2620: the three cards the random-policy soak caught resolving an
// ability whose source had ceased to exist or left its zone.

const (
	vanishedValleyMightcallerOracle = "16e9c452-6288-4da8-813d-2eb6b7a538c3"
	vanishedMageRingNetworkOracle   = "136596a0-b179-40be-b42d-c0b992621c95"
)

// vanishToken removes a token the way a dying one goes: destroyed, then
// ceasing to exist at the next state check (CR 704.5d).
func vanishToken(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Errorf("destroy the token: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if _, ok := g.LookupCardForEffect(id); ok {
		t.Fatal("test setup: the token should have ceased to exist")
	}
}

func TestTalonGatesOfMadaraLeavingHandInResponseDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	id := pushCatalogHandCard(me, "Talon Gates of Madara", "Land — Gate", talonGatesOfMadaraOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{4}: put this card from your hand onto the battlefield: %v", err)
	}
	// In response the card leaves the hand (a discard effect, say).
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(id); err != nil {
			t.Fatalf("exile from hand: %v", err)
		}
	})
	before := len(g.Events)
	passPriorityAroundTable(t, g)
	if n := countCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the ability should do nothing", n)
	}
	if g.Battlefield.Contains(id) {
		t.Error("a card that left the hand still entered the battlefield")
	}
}

func TestValleyMightcallerGoneBeforeTriggerResolvesDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	valley := pushCatalogPermanent(g, me.ID, "Valley Mightcaller", "Token Creature — Frog Warrior",
		vanishedValleyMightcallerOracle, false)

	castCatalogSpell(t, g, "Frog", "Creature — Frog", "", nil)
	for i := 0; i < 8 && triggerOnStack(g, valley) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if triggerOnStack(g, valley) == nil {
		t.Fatal("Valley Mightcaller's counter trigger never reached the stack")
	}
	vanishToken(t, g, valley)
	before := len(g.Events)
	passPriorityAroundTable(t, g)
	if n := countCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the trigger should do nothing", n)
	}
}

func TestMageRingNetworkGoneBeforeAbilityResolvesDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := pushCatalogPermanent(g, me.ID, "Mage-Ring Network", "Token Land",
		vanishedMageRingNetworkOracle, false)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("{1}, {T}: put a storage counter: %v", err)
	}
	vanishToken(t, g, land)
	before := len(g.Events)
	passPriorityAroundTable(t, g)
	if n := countCatalogEvents(g, game.EventEffectError, before); n != 0 {
		t.Errorf("EventEffectError count = %d, want 0 — the ability should do nothing", n)
	}
}
