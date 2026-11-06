package game

import (
	"testing"

	"github.com/google/uuid"
)

// trigger_condition_graveyard_test.go — FootprintCondGraveyard, the
// intervening "if" that reads its controller's graveyard (Walltop
// Sentries, Dragonfly Swarm: "if there's a Lesson card in your
// graveyard"). It conflicts with an item that can put a card into a
// graveyard or take one out, and with nothing else.

func TestAGraveyardConditionConflictsWithWhatMovesCards(t *testing.T) {
	f := newCondFixture(t)
	a := f.mustAccess(t, cond(FootprintCondGraveyard, 0), gainOne)
	if !a.graveyardReads {
		t.Fatal("a graveyard condition must read the graveyard")
	}
	if a.boardReads || len(a.reads) != 0 {
		t.Errorf("a graveyard condition read the board (%v) or objects (%v)", a.boardReads, a.reads)
	}
	for _, c := range []struct {
		name string
		b    FootprintStep
		want TriggerOrderReason
	}{
		{"a draw", drawOne, ""},
		{"a gain", gainOne, ""},
		{"a drain", FootprintStep{Kind: FootprintLoseLifeEachOpponent, N: 1}, ""},
		{"a scry", FootprintStep{Kind: FootprintScry, N: 1}, ""},
		{"energy", FootprintStep{Kind: FootprintEnergy, N: 1}, ""},
		{"a mill", FootprintStep{Kind: FootprintMill, N: 1}, TriggerOrderReasonCondition},
		{"a surveil", FootprintStep{Kind: FootprintSurveil, N: 1}, TriggerOrderReasonCondition},
		{"a counter on a permanent", FootprintStep{Kind: FootprintCounterOnSource, N: 1, Counter: CounterMinusOne}, TriggerOrderReasonCondition},
	} {
		t.Run(c.name, func(t *testing.T) {
			other := uuid.New()
			f.g.Battlefield.PushTop(Card{InstanceID: other, Name: "Other", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Owner: f.me.ID, Controller: f.me.ID})
			item := f.item()
			item.SourceCardID = other
			b := newTriggerAccess(item.Controller)
			if !f.g.addFootprintStepLocked(&b, item, c.b, nil) {
				t.Fatal("the check refused the other step")
			}
			f.g.markRulesLocked(&b)
			if r := f.conflict(a, b); r != c.want {
				t.Errorf("%q, want %q", r, c.want)
			}
		})
	}
}
