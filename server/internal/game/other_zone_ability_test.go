package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// other_zone_ability_test.go — #1221 / ADR 0020's 2026-09-22
// addendum: an activated ability that functions from a zone other
// than the battlefield and the hand, and the ExileSelf cost component
// that scavenge, embalm and eternalize pay.
//
// The ENGINE half only. The catalog half — the four keyword
// constructors and the three cards — is in
// cards/effects/graveyard_keywords_test.go; the enumerator's walk is
// in legal/other_zone_ability_test.go; the per-viewer wire split is in
// protocol/zone_ability_view_test.go.

// graveyardAbility is "{1}: <do nothing observable>" declared to
// function from a graveyard — the smallest ability that exercises the
// zone dimension without dragging a keyword's effect in.
func graveyardAbility(cost string) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: "Graveyard test " + cost,
		Cost:  AbilityCost{Mana: cost},
		Zones: []ZoneKind{ZoneGraveyard},
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}
}

// exileSelfAbility is scavenge's shape: a mana component plus "Exile
// this card from your graveyard".
func exileSelfAbility(cost string) ActivatedAbilityShape {
	return ActivatedAbilityShape{
		Label: "Exile-this test " + cost,
		Cost:  AbilityCost{Mana: cost, ExileSelf: true},
		Zones: []ZoneKind{ZoneGraveyard},
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}
}

// pushGraveyardCard puts a card carrying `abilities` into p's
// graveyard and returns its instance ID.
func pushGraveyardCard(g *Game, p *Player, name string, abilities ...ActivatedAbilityShape) uuid.UUID {
	c := NewCard(name, p.ID)
	c.TypeLine = "Creature — Test"
	c.Power, c.Toughness = 3, 3
	c.ActivatedAbilities = abilities
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// The zone dimension, one zone past the hand: a graveyard ability is
// activatable while the card sits in its owner's graveyard, and the
// activator's mana is really spent.
func TestActivateFromGraveyard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	card := pushGraveyardCard(g, me, "Reassembling Test", graveyardAbility("{1}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, card, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from graveyard: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want the {1} spent", me.ManaPool)
	}
	if !me.Graveyard.Contains(card) {
		t.Error("the source left the graveyard — this ability's cost does not move it")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("%d stack items, want the ability", len(g.StackMeta))
	}
}

// CR 113.6 is a two-way rule, and the graveyard half has to refuse
// exactly as the hand half does: the same ability on the BATTLEFIELD
// is not activatable, and costs nothing to find that out.
func TestGraveyardAbilityIsRefusedOnTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	c := NewCard("Reassembling Test", me.ID)
	c.TypeLine = "Creature — Test"
	c.ActivatedAbilities = []ActivatedAbilityShape{graveyardAbility("{1}")}
	g.Battlefield.PushTop(c)
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrActivationZoneNotAllowed) {
		t.Fatalf("err = %v, want ErrActivationZoneNotAllowed", err)
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the mana untouched — the refusal is before any payment", me.ManaPool)
	}
}

// CR 108.4: off the battlefield the card's OWNER is the "you" of its
// printed text, so another seat cannot activate an ability off a
// graveyard that is not theirs.
func TestGraveyardAbilityIsRefusedForANonOwner(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, them := g.Seats[0], g.Seats[1]
	card := pushGraveyardCard(g, me, "Reassembling Test", graveyardAbility("{1}"))
	them.ManaPool.AddMana(ManaToken{Color: "C"})

	err := g.ActivateCatalogAbility(them.ID, card, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrCardCallerMismatch) {
		t.Fatalf("err = %v, want ErrCardCallerMismatch", err)
	}
}

// The ExileSelf component: the source card is in exile once the
// activation returns, and the ability is on the stack above it.
func TestExileSelfCostExilesTheSourceAtAnnounce(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	card := pushGraveyardCard(g, me, "Scavenge Test", exileSelfAbility("{1}"))
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, card, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Graveyard.Contains(card) {
		t.Error("the source is still in the graveyard")
	}
	if !g.Exile.Contains(card) {
		t.Error("the source is not in exile")
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("%d stack items, want the ability — a cost that ends the source still announces", len(g.StackMeta))
	}
	// The whole reason no snapshot lives on the stack item: the
	// effect reads the card back by instance ID and finds it where
	// the cost put it, with its printed power intact.
	var found bool
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		found = ok && c.Power == 3
	})
	if !found {
		t.Error("the exiled source is not readable by instance ID with its printed power")
	}
}

// An ExileSelf ability whose source is NOT in a graveyard is refused
// before anything is paid — the sibling of DiscardSelf's hand check.
func TestExileSelfCostIsRefusedOutsideAGraveyard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	// Declared for BOTH zones so the CR 113.6 gate lets the
	// activation through and the cost's own check is what refuses it.
	ab := exileSelfAbility("{1}")
	ab.Zones = []ZoneKind{ZoneGraveyard, ZoneHand}
	card := pushHandCard(g, me, "Scavenge Test", ab)
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	err := g.ActivateCatalogAbility(me.ID, card, 0, ActivateAbilityParams{})
	if !errors.Is(err, ErrActivationZoneNotAllowed) {
		t.Fatalf("err = %v, want ErrActivationZoneNotAllowed", err)
	}
	if !me.Hand.Contains(card) {
		t.Error("the source moved on a refused activation")
	}
	if len(me.ManaPool) != 1 {
		t.Errorf("pool = %v, want the mana untouched", me.ManaPool)
	}
}

// An ability declared for EXILE is activatable from exile, which is
// the third zone the dimension opens and the one with no per-seat
// pile to hang the owner off.
func TestActivateFromExile(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	c := NewCard("Exile Test", me.ID)
	c.TypeLine = "Creature — Test"
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "Exile-zone test {1}",
		Cost:  AbilityCost{Mana: "{1}"},
		Zones: []ZoneKind{ZoneExile},
		Effect: func(g *Game, item *StackItem) error {
			return g.DrawNForEffect(item.Controller, 1)
		},
	}}
	g.Exile.PushTop(c)
	me.ManaPool.AddMana(ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, c.InstanceID, 0, ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate from exile: %v", err)
	}
	if len(g.StackMeta) != 1 {
		t.Fatalf("%d stack items, want the ability", len(g.StackMeta))
	}
}
