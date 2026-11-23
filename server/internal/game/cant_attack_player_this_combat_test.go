package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// cant_attack_player_this_combat_test.go — "they can't attack you this
// combat" (Champions of Minas Tirith, S82 GAP batch 118).

// A grant made in the beginning-of-combat step binds the declaration that
// follows, and only the player named.
func TestCantAttackPlayerThisCombatBindsTheCombatInProgress(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	protected, restricted, third := g.Seats[0], g.Seats[1], g.Seats[2]
	c := readyCreatureFor(g, restricted.ID, "Raider")
	advanceToDeclareOf(t, g, 1)
	// Back to the beginning of combat of the same turn is not possible, so
	// grant inside the declare step: it is still the combat phase.
	g.WithWriteLock(func() {
		g.GrantCantAttackPlayerThisCombatForEffect(restricted.ID, protected.ID, CantAttackScope{PlayerOnly: true}, "Test grant", uuid.Nil)
	})
	got := attackTargetsOf(g, c)
	if slices.Contains(got, protected.ID) {
		t.Fatalf("the protected player is still attackable: %v", got)
	}
	if !slices.Contains(got, third.ID) {
		t.Errorf("a third player must stay attackable: %v", got)
	}
	// It ends with the combat phase: their next turn is free.
	advanceToDeclareOf(t, g, 2)
	advanceToDeclareOf(t, g, 1)
	g.WithWriteLock(func() { findBattlefieldCard(g, c).SummonedThisTurn = false })
	if !slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Error("the restriction outlived its combat")
	}
}

// Outside a combat phase there is no combat to last through: nothing is
// registered (CR 611.2b).
func TestCantAttackPlayerThisCombatOutsideCombatDoesNothing(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	protected, restricted := g.Seats[0], g.Seats[1]
	advanceTo(t, g, StepPrecombatMain)
	g.WithWriteLock(func() {
		g.GrantCantAttackPlayerThisCombatForEffect(restricted.ID, protected.ID, CantAttackScope{}, "Test grant", uuid.Nil)
	})
	if got := len(g.Seats[1].Statics); got != 0 {
		t.Errorf("%d statics registered outside combat, want none", got)
	}
}
