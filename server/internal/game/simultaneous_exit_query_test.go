package game

import (
	"testing"

	"github.com/google/uuid"
)

func TestSameSimultaneousExitUsesBothMembersAndRestoresOuterSet(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	a := pushBattlefieldForTest(g, p.ID, "First", "Creature", "")
	b := pushBattlefieldForTest(g, p.ID, "Second", "Creature", "")
	c := pushBattlefieldForTest(g, p.ID, "Later", "Creature", "")
	g.WithWriteLock(func() {
		if g.InSameSimultaneousExitForEffect(a, b) {
			t.Fatal("no simultaneous exit is open")
		}
		closeOuter := g.beginSimultaneousExitLocked([]uuid.UUID{a, b})
		if !g.InSameSimultaneousExitForEffect(a, b) || g.InSameSimultaneousExitForEffect(a, c) {
			t.Fatal("both cards must belong to the same actual exit")
		}
		closeInner := g.beginSimultaneousExitLocked([]uuid.UUID{c})
		if g.InSameSimultaneousExitForEffect(a, b) {
			t.Fatal("an inner exit must replace the current exit set")
		}
		closeInner()
		if !g.InSameSimultaneousExitForEffect(a, b) || g.InSameSimultaneousExitForEffect(a, c) {
			t.Fatal("restoring an outer exit must still check both IDs")
		}
		closeOuter()
		if g.InSameSimultaneousExitForEffect(a, b) {
			t.Fatal("a closed exit must not leak membership")
		}
	})
}
