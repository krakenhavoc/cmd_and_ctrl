package game

import (
	"reflect"
	"testing"
)

// lose_subtypes_test.go — #4072, ModLoseSubtypes.

// Only the named subtype goes, whatever its case; card types, other
// subtypes and a later addition of the same subtype are handled by the
// timestamp order (CR 613.7).
func TestLoseSubtypesRemovesOnlyTheNamedSubtype(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Human Werewolf")
	registerScopedEffectForTest(t, g, id, []Mod{LoseSubtypesMod("human")}, IndefiniteDuration())
	c := scopedEffectChar(t, g, id)
	if !reflect.DeepEqual(c.Types, []string{"Creature"}) || !reflect.DeepEqual(c.Subtypes, []string{"Werewolf"}) {
		t.Errorf("type line = %v — %v, want Creature — Werewolf", c.Types, c.Subtypes)
	}
}

// An object without the subtype is unchanged, and the effect does not
// add one.
func TestLoseSubtypesOfASubtypeTheObjectLacksChangesNothing(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Bear")
	registerScopedEffectForTest(t, g, id, []Mod{LoseSubtypesMod("Human")}, IndefiniteDuration())
	if c := scopedEffectChar(t, g, id); !reflect.DeepEqual(c.Subtypes, []string{"Bear"}) {
		t.Errorf("subtypes = %v, want [Bear]", c.Subtypes)
	}
}

func TestLoseSubtypesModProblem(t *testing.T) {
	if loseSubtypesModProblem(LoseSubtypesMod("Human")) != "" {
		t.Error("a named subtype is sound")
	}
	if loseSubtypesModProblem(LoseSubtypesMod()) == "" {
		t.Error("no subtype is a newer binary's shape")
	}
	if loseSubtypesModProblem(LoseSubtypesMod("")) == "" {
		t.Error("an empty subtype is refused")
	}
	if loseSubtypesModProblem(AddSubtypesMod("Human")) != "" {
		t.Error("another kind is not this check's business")
	}
}
