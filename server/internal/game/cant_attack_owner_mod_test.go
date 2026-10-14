package game

import (
	"slices"
	"testing"
)

// cant_attack_owner_mod_test.go pins ADR 0180 decision 6: "they gain
// 'This creature can't attack its owner'" (Elrond of the White Council)
// as a scoped-effect mod. It writes ADR 0106 §2's NotOwner restriction,
// the owner read live, and it is the creature's own granted ability, so
// a later "loses all abilities" takes it and an earlier one does not.

func TestCantAttackOwnerModRefusesTheOwnerOnly(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner, other := g.Seats[0], g.Seats[1], g.Seats[2]
	stolen := pushOwnedByAnother(g, owner.ID, me.ID, "Borrowed Bear", "")
	registerScopedEffectForTest(t, g, stolen, []Mod{CantAttackOwnerMod()}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	got := attackTargetsOf(g, stolen)
	if slices.Contains(got, owner.ID) || !slices.Contains(got, other.ID) {
		t.Fatalf("targets = %v: want every opponent but the owner", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(stolen, owner.ID))
	if got, want := re.Sentence(), "Borrowed Bear can't attack its owner."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareAttacker(stolen, other.ID); err != nil {
		t.Fatalf("declare at another opponent: %v", err)
	}
}

func TestCantAttackOwnerModGoesWithALaterLossOfAbilities(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	stolen := pushOwnedByAnother(g, owner.ID, me.ID, "Borrowed Bear", "")
	registerScopedEffectForTest(t, g, stolen, []Mod{CantAttackOwnerMod()}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, stolen, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, stolen); !slices.Contains(got, owner.ID) {
		t.Fatalf("targets = %v: a later loss of all abilities takes the granted restriction", got)
	}
}

func TestCantAttackOwnerModSurvivesAnEarlierLossOfAbilities(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	stolen := pushOwnedByAnother(g, owner.ID, me.ID, "Borrowed Bear", "")
	registerScopedEffectForTest(t, g, stolen, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	registerScopedEffectForTest(t, g, stolen, []Mod{CantAttackOwnerMod()}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, stolen); slices.Contains(got, owner.ID) {
		t.Fatalf("targets = %v: an ability gained after the loss stays", got)
	}
}
