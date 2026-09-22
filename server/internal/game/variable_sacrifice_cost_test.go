package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// variable_sacrifice_cost_test.go — #1213: a sacrifice cost whose
// count the ACTIVATOR announces, in the two printed forms.
//
//	"Sacrifice one or more artifacts"  Radiant Lotus — an open count
//	"Sacrifice X Treasures"            Grim Hireling — the count is X
//
// #747 gave the clause a fixed count and effects.Register refused
// everything else. What this pins is the half that changed: how many
// permanents one payment may name, that the number is RECORDED on the
// announcement rather than recomputed at resolution, and that an
// announcement outside the bounds is refused with nothing paid. The
// catalog half is in cards/effects/variable_sacrifice_cards_test.go.

func varSacSpec(label string) *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: label,
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsArtifact()
		},
		Min: 1, Max: 1,
	}
}

// openSacrificeCost is "Sacrifice one or more artifacts": a floor of
// one with no printed ceiling.
func openSacrificeCost() AbilityCost {
	spec := varSacSpec("one or more artifacts")
	spec.Min, spec.Max = 1, 0
	return AbilityCost{SacrificeOther: spec}
}

// xSacrificeCost is "Sacrifice X artifacts": the count is the
// announced X, and the mana component deliberately has no {X} in it.
func xSacrificeCost() AbilityCost {
	spec := varSacSpec("X artifacts")
	spec.CountFromX = true
	return AbilityCost{SacrificeOther: spec}
}

// pushVarSacSource seats an artifact whose ability records how many
// permanents the announcement actually paid, so a test can read the
// number the EFFECT sees rather than the number the board shows.
func pushVarSacSource(g *Game, owner *Player, cost AbilityCost, paid *int) uuid.UUID {
	c := NewCard("Variable Outlet", owner.ID)
	c.TypeLine = "Artifact"
	c.Controller = owner.ID
	c.ActivatedAbilities = []ActivatedAbilityShape{{
		Label: "sacrifice: record",
		Cost:  cost,
		Effect: func(_ *Game, item *StackItem) error {
			*paid = item.Paid.Sacrificed
			return nil
		},
	}}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func pushVarSacArtifact(g *Game, controller *Player, name string) uuid.UUID {
	c := NewCard(name, controller.ID)
	c.TypeLine = "Artifact"
	c.Controller = controller.ID
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// An "Activate only once each turn"-free ability whose count is the
// activator's: naming three pays three, and the EFFECT reads three
// back out of the announcement after the artifacts are gone.
func TestOpenSacrificeCountIsAnnouncedAndRecorded(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	src := pushVarSacSource(g, me, openSacrificeCost(), &seen)
	a := pushVarSacArtifact(g, me, "Rock A")
	b := pushVarSacArtifact(g, me, "Rock B")
	c := pushVarSacArtifact(g, me, "Rock C")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b, c},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if g.Battlefield.Contains(id) {
			t.Error("a named artifact survived the announcement")
		}
	}
	passBothForTest(g)
	if seen != 3 {
		t.Errorf("the effect read %d sacrificed, want 3 — the count is a fact about the announcement", seen)
	}
}

// The same ability with a different announcement: one is a legal
// payment too, and the record says one.
func TestOpenSacrificeAcceptsItsFloor(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	src := pushVarSacSource(g, me, openSacrificeCost(), &seen)
	a := pushVarSacArtifact(g, me, "Rock A")
	b := pushVarSacArtifact(g, me, "Rock B")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !g.Battlefield.Contains(b) {
		t.Error("an artifact nobody named was sacrificed")
	}
	passBothForTest(g)
	if seen != 1 {
		t.Errorf("the effect read %d sacrificed, want 1", seen)
	}
}

// Naming nothing is not a payment: the floor is one, and CR 118.3
// refuses the whole announcement rather than making the ability free.
func TestOpenSacrificeRefusesAnEmptyPayment(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	src := pushVarSacSource(g, me, openSacrificeCost(), &seen)
	rock := pushVarSacArtifact(g, me, "Rock")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("a refused activation sacrificed something")
	}
	if len(g.StackMeta) != 0 {
		t.Errorf("StackMeta has %d items, want 0", len(g.StackMeta))
	}
}

// "Sacrifice X artifacts": the announced X IS the count, so naming a
// different number of permanents is a refused announcement rather
// than a cheap one.
func TestSacrificeXMatchesTheAnnouncedX(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	src := pushVarSacSource(g, me, xSacrificeCost(), &seen)
	a := pushVarSacArtifact(g, me, "Rock A")
	b := pushVarSacArtifact(g, me, "Rock B")

	// X = 2 with one artifact named.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: []uuid.UUID{a},
	}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming one for X=2: err = %v, want ErrInvalidParam", err)
	}
	// X = 1 with two named.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		XValue:       1,
		SacrificeIDs: []uuid.UUID{a, b},
	}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("naming two for X=1: err = %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("a refused announcement sacrificed something")
	}
	// X = 2 with two named is the payment.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	passBothForTest(g)
	if seen != 2 {
		t.Errorf("the effect read %d sacrificed, want 2", seen)
	}
}

// A CountFromX sacrifice clause makes the ability demand an X even
// though its mana component has none, and it adds NO generic mana to
// the cost — Grim Hireling's {B} is {B} at every X.
func TestSacrificeXDemandsXWithoutChangingTheManaCost(t *testing.T) {
	cost := xSacrificeCost()
	cost.Mana = "{B}"
	if !cost.DemandsX() {
		t.Error("DemandsX is false for a cost whose sacrifice clause counts from X")
	}
	if got := cost.XSlots(); got != 0 {
		t.Errorf("XSlots = %d, want 0 — a sacrifice clause's X buys permanents, not mana", got)
	}
	plain := AbilityCost{Mana: "{B}"}
	if plain.DemandsX() {
		t.Error("DemandsX is true for a cost with no X anywhere")
	}
}

// The floor still applies with a variable clause: a board that cannot
// reach it cannot activate at all.
func TestOpenSacrificeRefusedWithAnEmptyBoard(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	src := pushVarSacSource(g, me, openSacrificeCost(), &seen)

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if SacrificeCountLegal(openSacrificeCost().SacrificeOther, 0, 0) {
		t.Error("naming nothing satisfies a floor of one")
	}
}

// A FIXED clause is untouched by all of this: the same validator, the
// same count, the same refusals it had before #1213.
func TestFixedSacrificeCountIsUnchanged(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me := g.Seats[0]
	var seen int
	spec := varSacSpec("two artifacts")
	spec.Min, spec.Max = 2, 2
	src := pushVarSacSource(g, me, AbilityCost{SacrificeOther: spec}, &seen)
	a := pushVarSacArtifact(g, me, "Rock A")
	b := pushVarSacArtifact(g, me, "Rock B")
	c := pushVarSacArtifact(g, me, "Rock C")

	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a},
	}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("one for a two-permanent clause: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b, c},
	}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("three for a two-permanent clause: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passBothForTest(g)
	if seen != 2 {
		t.Errorf("the effect read %d sacrificed, want 2", seen)
	}
}
