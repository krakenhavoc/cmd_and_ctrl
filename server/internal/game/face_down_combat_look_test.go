package game

import (
	"slices"
	"testing"

	"github.com/google/uuid"
)

// S82 old b334: Revealing Wind's one-shot look at the face-down creatures in
// combat (face_down_combat_look.go).
func TestLookAtFaceDownCombatantsShowsOnlyThoseInCombat(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	attacker := pushLookTestFaceDown(t, g, opp.ID, "Attacking Morph")
	idle := pushLookTestFaceDown(t, g, opp.ID, "Idle Morph")
	mine := pushLookTestFaceDown(t, g, me.ID, "My Morph")
	faceUp := pushLookTestPermanent(g, opp.ID, "Grizzly Bears", "Creature — Bear", "")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{attacker, mine, faceUp} {
			g.Battlefield.Cards[findCardOnBattlefield(g, id)].AttackingTarget = me.ID
		}
	})
	var seen []uuid.UUID
	g.WithWriteLock(func() { seen = g.LookAtFaceDownCombatantsForEffect(me.ID) })
	if !slices.Equal(seen, []uuid.UUID{attacker}) {
		t.Fatalf("looked at %v, want only the opponent's face-down attacker", seen)
	}
	c, _ := g.LookupCardForEffect(attacker)
	if !c.IsKnownTo(me.ID) || c.IsKnownTo(other.ID) {
		t.Errorf("knowers: me %v other %v, want me alone beside the controller", c.IsKnownTo(me.ID), c.IsKnownTo(other.ID))
	}
	c, _ = g.LookupCardForEffect(idle)
	if c.IsKnownTo(me.ID) {
		t.Error("a face-down creature not in combat was shown")
	}
}
