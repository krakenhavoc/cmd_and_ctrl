package game

import "testing"

// assigns_as_unblocked_mod_test.go — ModAssignsAsThoughUnblocked (Garruk,
// Savage Herald's -7).

// The grant sets the flag the combat damage step reads, and only on the
// affected object.
func TestAssignsAsThoughUnblockedModSetsTheFlag(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Beast")
	other := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Beast")
	if scopedEffectChar(t, g, id).AssignsAsThoughUnblocked {
		t.Fatal("control: the flag is set before the effect")
	}
	registerScopedEffectForTest(t, g, id, []Mod{AssignsAsThoughUnblockedMod()}, IndefiniteDuration())
	if !scopedEffectChar(t, g, id).AssignsAsThoughUnblocked {
		t.Error("the affected creature has no flag")
	}
	if scopedEffectChar(t, g, other).AssignsAsThoughUnblocked {
		t.Error("an unaffected creature has the flag")
	}
}

// The mandatory grant (Outmaneuver) sets its own flag, not the optional one.
func TestMustAssignAsThoughUnblockedModSetsItsOwnFlag(t *testing.T) {
	g := newActiveGame(t)
	id := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Beast")
	other := pushTypedTestPermanent(g, g.Seats[0].ID, "Creature — Beast")
	registerScopedEffectForTest(t, g, id, []Mod{MustAssignAsThoughUnblockedMod()}, IndefiniteDuration())
	ch := scopedEffectChar(t, g, id)
	if !ch.MustAssignAsThoughUnblocked || ch.AssignsAsThoughUnblocked {
		t.Errorf("flags on the affected creature: must %v, may %v; want must only", ch.MustAssignAsThoughUnblocked, ch.AssignsAsThoughUnblocked)
	}
	if scopedEffectChar(t, g, other).MustAssignAsThoughUnblocked {
		t.Error("an unaffected creature has the flag")
	}
}
