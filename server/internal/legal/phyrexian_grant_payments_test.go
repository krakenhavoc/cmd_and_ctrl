package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// phyrexian_grant_payments_test.go — ADR 0131 §2 (#2531), PR 2: the
// enumerator offers the life answer for the two payments PR 1 left on
// mana alone — a mana ability's own mana component, and a mana
// pay_unless — under K'rrik, Son of Yawgmoth or for a printed {B/P}. The
// counts are the ones a cast gets, bounded by the engine's own life
// predicate, and every move offered dispatches.

// filterLand seats a "{B}, {T}: Add {B}{B}" source under `p`.
func filterLand(g *game.Game, p *game.Player, cost string) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: "Black Filter", TypeLine: "Land",
		ManaAbilities: []game.ManaAbilityShape{{
			TapCost: true, ManaCost: cost, Produced: "{B}{B}", Label: cost + ", {T}: Add {B}{B}",
		}},
	})
}

func manaMovesOf(moves []legal.Move, source uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range moves {
		if m.Kind == legal.KindMana && m.Source == source {
			out = append(out, m)
		}
	}
	return out
}

// With nothing to pay the {B} but life, the filter is offered at the
// price of 2 life under K'rrik, and not offered without it.
func TestKrrikOffersAManaAbilitysBlackSymbolForLife(t *testing.T) {
	for _, withKrrik := range []bool{true, false} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		src := filterLand(g, seat, "{B}")
		if withKrrik {
			krrikOn(g, seat)
		}
		advanceTo(t, g, game.StepPrecombatMain)
		got := manaMovesOf(legal.EnumerateFor(g, seat.ID), src)
		if !withKrrik {
			if len(got) != 0 {
				t.Fatalf("filter offered with no black source and no K'rrik: %v", labels(got))
			}
			continue
		}
		if len(got) != 1 {
			t.Fatalf("filter moves = %d, want 1: %v", len(got), labels(got))
		}
		if n := phyrexianLifeOf(t, got[0]); n != 1 {
			t.Errorf("phyrexian_life = %d, want 1", n)
		}
		if got[0].Cost == nil || got[0].Cost.Life != 2 || got[0].Cost.PhyrexianLife != 2 {
			t.Errorf("cost = %+v, want Life = PhyrexianLife = 2", got[0].Cost)
		}
		before := seat.Life
		dispatchOne(t, g, seat.ID, got[0])
		if seat.Life != before-2 {
			t.Errorf("life after the activation = %d, want %d", seat.Life, before-2)
		}
	}
}

// Mana first: with a Swamp to tap the filter carries no life.
func TestKrrikManaAbilityIsPaidInManaWhenTheBoardHasIt(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	src := filterLand(g, seat, "{B}")
	battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	krrikOn(g, seat)
	advanceTo(t, g, game.StepPrecombatMain)
	got := manaMovesOf(legal.EnumerateFor(g, seat.ID), src)
	if len(got) != 1 {
		t.Fatalf("filter moves = %d, want 1: %v", len(got), labels(got))
	}
	if n := phyrexianLifeOf(t, got[0]); n != 0 {
		t.Errorf("phyrexian_life = %d with a Swamp to tap, want 0", n)
	}
	dispatchAll(t, g, seat.ID, got)
}

// The bound is the engine's: at 1 life the claim is not offered.
func TestKrrikManaAbilityLifeIsBoundedByTheLifeTotal(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	src := filterLand(g, seat, "{B}")
	krrikOn(g, seat)
	advanceTo(t, g, game.StepPrecombatMain)
	seat.Life = 1
	if got := manaMovesOf(legal.EnumerateFor(g, seat.ID), src); len(got) != 0 {
		t.Fatalf("filter offered at 1 life: %v", labels(got))
	}
}

// A printed {B/P} on a mana ability needs no grant.
func TestPrintedPhyrexianManaAbilityIsOfferedForLife(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	src := filterLand(g, seat, "{B/P}")
	advanceTo(t, g, game.StepPrecombatMain)
	got := manaMovesOf(legal.EnumerateFor(g, seat.ID), src)
	if len(got) != 1 || phyrexianLifeOf(t, got[0]) != 1 {
		t.Fatalf("filter moves = %v, want one paid with 1 symbol of life", labels(got))
	}
	dispatchAll(t, g, seat.ID, got)
}

// --- pay_unless ------------------------------------------------------------

func payUnlessChoices(t *testing.T, g *game.Game, payer *game.Player, cost string) []legal.Move {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.QueuePayUnlessForEffect(payer.ID, uuid.New(), cost,
			"Pay "+cost+"?", func(*game.Game) error { return nil }); err != nil {
			t.Fatalf("QueuePayUnlessForEffect: %v", err)
		}
	})
	var out []legal.Move
	for _, m := range legal.EnumerateFor(g, payer.ID) {
		if m.Kind == legal.KindChoice {
			out = append(out, m)
		}
	}
	return out
}

// "Unless that player pays {B}" with no black mana: under K'rrik the pay
// answer is offered at 2 life, carried in `phyrexian_life`. Without it
// the decline stands alone. Every answer dispatches.
func TestKrrikOffersAPayUnlessPaidWithLife(t *testing.T) {
	for _, withKrrik := range []bool{true, false} {
		g := newTable(t)
		payer := g.Seats[1]
		if withKrrik {
			krrikOn(g, payer)
		}
		choices := payUnlessChoices(t, g, payer, "{B}")
		want := 1
		if withKrrik {
			want = 2
		}
		if len(choices) != want {
			t.Fatalf("K'rrik %v: %d answers %v, want %d", withKrrik, len(choices), labels(choices), want)
		}
		dispatchAll(t, g, payer.ID, choices)
		if !withKrrik {
			continue
		}
		var pay *legal.Move
		for i := range choices {
			if phyrexianLifeOf(t, choices[i]) > 0 {
				pay = &choices[i]
			}
		}
		if pay == nil {
			t.Fatalf("no pay answer carried phyrexian_life: %v", labels(choices))
		}
		if phyrexianLifeOf(t, *pay) != 1 || pay.Cost == nil || pay.Cost.Life != 2 || pay.Cost.PhyrexianLife != 2 {
			t.Errorf("pay answer %q cost %+v phyrexian_life %d, want Life = PhyrexianLife = 2 and 1 symbol",
				pay.Label, pay.Cost, phyrexianLifeOf(t, *pay))
		}
		before := payer.Life
		dispatchOne(t, g, payer.ID, *pay)
		if payer.Life != before-2 {
			t.Errorf("life after paying = %d, want %d", payer.Life, before-2)
		}
	}
}

// Mana first: a Swamp pays the {B} and the answer carries no life.
func TestKrrikPayUnlessIsPaidInManaWhenTheBoardHasIt(t *testing.T) {
	g := newTable(t)
	payer := g.Seats[1]
	krrikOn(g, payer)
	battlefieldCard(g, payer, basic("Swamp", "Swamp"))
	choices := payUnlessChoices(t, g, payer, "{B}")
	if len(choices) != 2 {
		t.Fatalf("%d answers %v, want pay and decline", len(choices), labels(choices))
	}
	for _, m := range choices {
		if n := phyrexianLifeOf(t, m); n != 0 {
			t.Errorf("%q carries phyrexian_life %d with a Swamp to tap", m.Label, n)
		}
	}
	dispatchAll(t, g, payer.ID, choices)
}

// At 1 life the pay answer is not offered (CR 119.4).
func TestKrrikPayUnlessLifeIsBoundedByTheLifeTotal(t *testing.T) {
	g := newTable(t)
	payer := g.Seats[1]
	krrikOn(g, payer)
	payer.Life = 1
	choices := payUnlessChoices(t, g, payer, "{B}")
	if len(choices) != 1 {
		t.Fatalf("%d answers %v at 1 life, want the decline alone", len(choices), labels(choices))
	}
}
