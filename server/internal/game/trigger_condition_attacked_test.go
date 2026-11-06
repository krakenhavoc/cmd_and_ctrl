package game

import (
	"testing"

	"github.com/google/uuid"
)

// Raid's "if you attacked this turn" (FootprintCondAttacked) counts
// declared attackers, which only a turn-based action adds: it reads
// nothing while it holds, and a step that would read it before anyone
// attacked is refused, since a later declaration could make it true.
func TestAttackedConditionReadsNothingWhileItHolds(t *testing.T) {
	f := newCondFixture(t)
	if _, ok := f.access(t, cond(FootprintCondAttacked, 0), drawOne); ok {
		t.Error("before an attack the condition is false, and the step is refused")
	}
	f.g.TurnTally.Players = map[uuid.UUID]PlayerTurnTally{f.me.ID: {AttacksDeclared: 1}}
	a := f.mustAccess(t, cond(FootprintCondAttacked, 0), drawOne)
	if a.boardReads || a.monarchRead {
		t.Errorf("read the board or the monarch: %+v", a)
	}
}
