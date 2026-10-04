package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adds_no_mana_now_test.go — ADR 0117 §5. game.ManaAbilityAddsNoMana
// also means "adds nothing right now" for a COMPUTED output: a
// power-0 Vivi Ornitier, a Selvala at greatest power 0, an Exotic
// Orchard with nothing to copy. The engine still accepts the
// activation (CR 605.1a); what changes is that nothing OFFERS it.

func addsNoManaNow(g *game.Game, controller, card uuid.UUID, idx int) bool {
	var out bool
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == card {
				out = game.ManaAbilityAddsNoMana(g, controller, card, game.ManaAbilitiesForCard(g.Battlefield.Cards[i])[idx])
			}
		}
	})
	return out
}

func TestAddsNoManaFollowsViviPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	advanceToMain(t, g)

	if !addsNoManaNow(g, me.ID, vivi, 0) {
		t.Error("power-0 Vivi: adds_no_mana = false, want true")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, 2) })
	if addsNoManaNow(g, me.ID, vivi, 0) {
		t.Error("power-2 Vivi: adds_no_mana = true, want false")
	}
}

func TestAddsNoManaFollowsSelvalaGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	selvala := b12Push(g, me.ID, "Selvala, Heart of the Wilds", "Legendary Creature — Elf Scout", selvalaHeartOracle, 0, 3)
	advanceToMain(t, g)
	if !addsNoManaNow(g, me.ID, selvala, 0) {
		t.Error("Selvala at greatest power 0: adds_no_mana = false, want true")
	}
	b16Creature(g, me.ID, "Big Friend", "Creature — Beast", 4, 4, "G")
	if addsNoManaNow(g, me.ID, selvala, 0) {
		t.Error("Selvala at greatest power 4: adds_no_mana = true, want false")
	}
}

func TestAddsNoManaFollowsExoticOrchardMatch(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	orchard := seedPermanentWithOracle(g, me.ID, "Exotic Orchard", "Land", exoticOrchardOracle)
	if !addsNoManaNow(g, me.ID, orchard, 0) {
		t.Error("Exotic Orchard with no opposing lands: adds_no_mana = false, want true")
	}
	seedManaLand(g, them.ID, "Island", "Basic Land — Island", "U")
	if addsNoManaNow(g, me.ID, orchard, 0) {
		t.Error("Exotic Orchard with an opposing Island: adds_no_mana = true, want false")
	}
}

// A static output is not this rule's business: fixed production never
// reads as adding nothing.
func TestAddsNoManaIgnoresAFixedOutput(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	forest := seedManaLand(g, me.ID, "Forest", "Basic Land — Forest", "G")
	if addsNoManaNow(g, me.ID, forest, 0) {
		t.Error("a Forest reads as adding no mana")
	}
}

// The view: the zero case ships adds_no_mana and no color_options;
// the same Vivi at power 2 ships color_options and no flag.
func TestViewShipsAddsNoManaForAPowerZeroVivi(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	advanceToMain(t, g)

	row := upfrontManaRow(t, g, me.ID, vivi, 0)
	if !row.AddsNoMana || row.ColorOptions != nil {
		t.Errorf("power 0: adds_no_mana=%v color_options=%v, want true and none", row.AddsNoMana, row.ColorOptions)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(vivi, game.CounterPlusOne, 2) })
	row2 := upfrontManaRow(t, g, me.ID, vivi, 0)
	if row2.AddsNoMana || len(row2.ColorOptions) != 2 {
		t.Errorf("power 2: adds_no_mana=%v color_options=%v, want false and two lists", row2.AddsNoMana, row2.ColorOptions)
	}
}

// The engine still accepts the activation at power 0: it adds nothing.
func TestActivationOfAnAddsNothingAbilityIsStillAccepted(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	vivi := b12Push(g, me.ID, "Vivi Ornitier", "Legendary Creature — Wizard", viviOrnitierOracle, 0, 3)
	advanceToMain(t, g)
	if !addsNoManaNow(g, me.ID, vivi, 0) {
		t.Fatal("precondition: power-0 Vivi should read as adding no mana")
	}
	if err := g.ActivateManaAbility(me.ID, vivi, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility refused a power-0 Vivi: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool has %d tokens, want 0", len(me.ManaPool))
	}
}
