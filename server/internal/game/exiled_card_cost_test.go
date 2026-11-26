package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exiled_card_cost_test.go — #6270, Back from the Brink: "Exile a creature
// card from your graveyard and pay its mana cost: ...". CR 602.2b: the
// card is named as the cost is announced, and the total cost (CR 601.2f)
// is then its mana cost; CR 601.2h: the exile and the mana are paid
// together, all or nothing.

func brinkSource(g *Game, me *Player) uuid.UUID {
	return abilitySource(g, me, AbilityCost{
		ExileCards:         &ExileCost{N: 1, Label: "a creature card", From: ZoneGraveyard, Match: func(c Card) bool { return c.IsCreature() }},
		ManaFromExiledCard: true,
	})
}

func graveyardCreature(p *Player, name, cost string) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine, c.ManaCost = "Creature — Beast", cost
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// A {2}{G} creature card costs {2}{G} with the exile.
func TestExileAndPayItsManaCostChargesTheCardsCost(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := brinkSource(g, me)
	beast := graveyardCreature(me, "Beast", "{2}{G}")
	me.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ExileIDs: []uuid.UUID{beast}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Exile.Contains(beast) {
		t.Error("the creature card was not exiled")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want the {2}{G} spent", n)
	}
}

// A card the pool can't pay for can't be named: nothing is exiled and no
// mana is spent (CR 601.2h, no partial payment).
func TestExileAndPayItsManaCostRefusesAnUnaffordableCard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := brinkSource(g, me)
	big := graveyardCreature(me, "Big Beast", "{4}{G}")
	me.ManaPool.AddMana(ManaToken{Color: "G"}, ManaToken{Color: "C"}, ManaToken{Color: "C"})
	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{ExileIDs: []uuid.UUID{big}, Strict: true})
	if err == nil {
		t.Fatal("a {4}{G} card was exiled for three mana")
	}
	if !me.Graveyard.Contains(big) || me.ManaPool.Total() != 3 {
		t.Error("a refused activation exiled the card or spent mana")
	}
}

// With no creature card in the graveyard it can't be activated.
func TestExileAndPayItsManaCostNeedsACreatureCard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	src := brinkSource(g, me)
	seedGraveyard(me, "Spell", "Sorcery", "{G}")
	err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{Strict: true})
	if !errors.Is(err, ErrInvalidParam) {
		t.Fatalf("activate with no creature card: %v, want ErrInvalidParam", err)
	}
}
