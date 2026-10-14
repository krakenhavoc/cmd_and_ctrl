package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestAnAnyDefenderTaxIsEnumeratedAtItsPrice — #3456: War Tax's record
// taxes an attack on every seat, and the enumerator prices it with the
// engine's own pricer, offering only what the lands can pay.
func TestAnAnyDefenderTaxIsEnumeratedAtItsPrice(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)

	freshCreature(g, me, "Bear")
	untappedIslands(g, me, 2)
	g.WithWriteLock(func() {
		g.AnyAttackTaxForEffect(uuid.Nil, "{2}", game.Duration{}, "War Tax test")
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)

	n := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack {
			continue
		}
		n++
		if m.Cost == nil || m.Cost.Mana != "{2}" {
			t.Errorf("attack %q costs %#v, want {2}", m.Label, m.Cost)
		}
	}
	if n == 0 {
		t.Fatalf("two lands cover {2} but no attack was offered: %v", labels(moves))
	}
}

// TestAnUnaffordableAnyDefenderTaxHidesEveryAttack — with no mana the
// attack on any seat is not offered.
func TestAnUnaffordableAnyDefenderTaxHidesEveryAttack(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)

	freshCreature(g, me, "Bear")
	g.WithWriteLock(func() {
		g.AnyAttackTaxForEffect(uuid.Nil, "{2}", game.Duration{}, "War Tax test")
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	for _, m := range moves {
		if m.Kind == legal.KindAttack {
			t.Errorf("offered an unaffordable attack: %q", m.Label)
		}
	}
}
