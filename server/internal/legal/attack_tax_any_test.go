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

// TestALifeTaxIsOfferedAtItsPriceAndHiddenWhenUnpayable — #5869:
// Sivitri's "pays 2 life for each of those creatures" is a move cost in
// life, and an attack the seat's life total can't pay is not offered
// (CR 119.4).
func TestALifeTaxIsOfferedAtItsPriceAndHiddenWhenUnpayable(t *testing.T) {
	for _, tc := range []struct {
		life    int
		offered bool
	}{{2, true}, {1, false}} {
		g := newTable(t)
		me := g.Seats[g.Turn.ActiveSeat]
		clearHand(me)
		freshCreature(g, me, "Bear")
		var defenders []uuid.UUID
		for _, p := range g.Seats {
			if p.ID != me.ID {
				defenders = append(defenders, p.ID)
			}
		}
		g.WithWriteLock(func() {
			for _, d := range defenders {
				g.AttackTaxLifeForEffect(uuid.Nil, d, 2, game.AttackTaxOnPlayerOrPlaneswalkers, game.Duration{}, "Sivitri test")
			}
		})
		advanceTo(t, g, game.StepDeclareAttackers)
		me.Life = tc.life

		moves := legal.EnumerateFor(g, me.ID)
		n := 0
		for _, m := range moves {
			if m.Kind != legal.KindAttack {
				continue
			}
			n++
			if m.Cost == nil || m.Cost.Life != 2 || m.Cost.Mana != "" {
				t.Errorf("life %d: attack %q costs %#v, want 2 life", tc.life, m.Label, m.Cost)
			}
		}
		if tc.offered && n == 0 {
			t.Errorf("life %d: no attack offered: %v", tc.life, labels(moves))
		}
		if !tc.offered && n != 0 {
			t.Errorf("life %d: %d unpayable attacks offered", tc.life, n)
		}
	}
}
