package legal_test

import (
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// phyrexian_cast_life_test.go — #1677. A spell's Phyrexian symbols
// can be paid with 2 life each (CR 107.4f), announced on the cast as
// `phyrexian_life` (CR 601.2b). The enumerator used to price the mana
// path only, so a bot with one Swamp was never offered Dismember at any
// life total. It now offers the mana payment when the seat has it, the
// fewest life symbols that make the cast affordable when it does not,
// and one all-life payment — each bounded by the life total, and each
// one the engine accepts (every test dispatches everything it is
// offered).

// dismemberCost is Dismember's printed cost: two Phyrexian black
// symbols and one generic.
const dismemberCost = "{1}{B/P}{B/P}"

// phyrexianDismember seats a Dismember-shaped instant in hand on a
// main phase, with `swamps` Swamps and the seat at `life`. Not the
// catalog card — the cost is what is under test, and a target clause
// would only multiply the moves.
func phyrexianDismember(t *testing.T, swamps, life int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	seat.Life = life
	for i := 0; i < swamps; i++ {
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	}
	card := handCard(seat, game.Card{
		Name:     "Dismember",
		TypeLine: "Instant",
		ManaCost: dismemberCost,
		Layout:   "normal",
	})
	return g, seat, card
}

// lifeCounts is the sorted phyrexian_life of every cast offered for
// `card`, checking on the way that the move's declared cost agrees
// with its params: Cost.Life and Cost.PhyrexianLife are both 2 per
// symbol, and a mana payment declares no cost at all.
func lifeCounts(t *testing.T, moves []legal.Move, card uuid.UUID) []int {
	t.Helper()
	var out []int
	for _, m := range castMovesFor(moves, card) {
		n := phyrexianLifeOf(t, m)
		out = append(out, n)
		points := n * game.PhyrexianLifePerSymbol
		switch {
		case n == 0 && m.Cost != nil:
			t.Errorf("mana payment %q declares a cost %+v", m.Label, *m.Cost)
		case n > 0 && (m.Cost == nil || m.Cost.Life != points || m.Cost.PhyrexianLife != points):
			t.Errorf("%q pays %d symbols with life but declares %+v, want Life = PhyrexianLife = %d", m.Label, n, m.Cost, points)
		}
	}
	sort.Ints(out)
	return out
}

func sameInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestPhyrexianCastPaymentsOffered is the table: which counts are
// offered for {1}{B/P}{B/P} by Swamps and life total, and every one
// of them dispatches.
func TestPhyrexianCastPaymentsOffered(t *testing.T) {
	for _, tc := range []struct {
		name   string
		swamps int
		life   int
		want   []int
	}{
		// The issue's case: one Swamp pays the {1}, and both symbols
		// have to be bought with 4 life.
		{"one Swamp at 20 life", 1, 20, []int{2}},
		// Two Swamps: the fewest symbols by life is one, and the
		// all-life payment rides beside it.
		{"two Swamps at 20 life", 2, 20, []int{1, 2}},
		// Three Swamps pay it all: the mana payment, and the free
		// cast that saves the Swamps for something else.
		{"three Swamps at 20 life", 3, 20, []int{0, 2}},
		// CR 119.4: 3 life cannot pay 4. One Swamp needs both symbols
		// by life, so there is no cast at all.
		{"one Swamp at 3 life", 1, 3, nil},
		// Two Swamps at 3 life: one symbol for 2 life is payable,
		// the 4-life payment is not offered.
		{"two Swamps at 3 life", 2, 3, []int{1}},
		// Exactly 4 life may be paid (CR 119.4 is "down to 0"): the
		// enumerator offers what the engine accepts; whether to take
		// it is the policy's call.
		{"one Swamp at 4 life", 1, 4, []int{2}},
		// No Swamps: the {1} has no source whatever the life.
		{"no Swamps at 20 life", 0, 20, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, seat, card := phyrexianDismember(t, tc.swamps, tc.life)
			moves := legal.EnumerateFor(g, seat.ID)
			if got := lifeCounts(t, moves, card); !sameInts(got, tc.want) {
				t.Fatalf("phyrexian_life offered = %v, want %v", got, tc.want)
			}
			dispatchAll(t, g, seat.ID, castMovesFor(moves, card))
		})
	}
}

// TestPhyrexianLifeCastPaysTheLife — the offered 4-life Dismember is
// not only accepted, it charges what it said: the seat loses 4 life
// and the Swamp taps for the {1}.
func TestPhyrexianLifeCastPaysTheLife(t *testing.T) {
	g, seat, card := phyrexianDismember(t, 1, 20)
	got := castMovesFor(legal.EnumerateFor(g, seat.ID), card)
	if len(got) != 1 {
		t.Fatalf("offered %d casts, want the one 4-life payment", len(got))
	}
	dispatchAll(t, g, seat.ID, got)
	// dispatchAll plays each move on a clone; play this one for real.
	dispatchOne(t, g, seat.ID, got[0])
	if seat.Life != 16 {
		t.Errorf("life after the cast = %d, want 16", seat.Life)
	}
	if !g.Stack.Contains(card) {
		t.Error("Dismember is not on the stack")
	}
}

// TestPhyrexianCastLifeIsNotOfferedToALockedSeat — CR 119.8: a
// player whose life total can't change can't pay life, so the
// strike refuses any claim and the enumerator must not offer one
// (the #1200 shape, on the cast path).
func TestPhyrexianCastLifeIsNotOfferedToALockedSeat(t *testing.T) {
	g, seat, card := phyrexianDismember(t, 1, 20)
	g.WithWriteLock(func() {
		g.GrantLifeTotalLockForEffect(seat.ID, "Test — your life total can't change",
			uuid.Nil, g.UntilYourNextTurnDuration(seat.ID))
	})
	if got := lifeCounts(t, legal.EnumerateFor(g, seat.ID), card); len(got) != 0 {
		t.Errorf("a locked seat was offered phyrexian_life %v", got)
	}
}

// TestGrantedPhyrexianCastPaysWithLife — Gonti's grant (#1589 /
// #1676): a stolen {1}{B/P}{B/P} cast from exile with "mana of any
// type can be spent" keeps its life option, and the enumerator offers
// it off one Mountain. Under #1676's pricing the symbols stay
// Phyrexian (AnyMana), so the strike finds them.
func TestGrantedPhyrexianCastPaysWithLife(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lands int
		want  []int
	}{
		{"one Mountain", 1, []int{2}},
		{"three Mountains", 3, []int{0, 2}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newTable(t)
			seat := g.Seats[g.Turn.ActiveSeat]
			victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			clearHand(seat)
			advanceTo(t, g, game.StepPrecombatMain)
			for i := 0; i < tc.lands; i++ {
				battlefieldCard(g, seat, basic("Mountain", "Mountain"))
			}
			loot := game.NewCard("Stolen Dismember", victim.ID)
			loot.TypeLine = "Sorcery"
			loot.ManaCost = dismemberCost
			victim.Library.PushTop(loot)
			g.WithWriteLock(func() {
				if _, err := g.ExileTopWithPermissionForEffect(victim.ID, seat.ID, 1,
					game.CastPermission{AnyType: true, Duration: game.WhileInZoneDuration()}); err != nil {
					t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
				}
			})
			moves := legal.EnumerateFor(g, seat.ID)
			if got := lifeCounts(t, moves, loot.InstanceID); !sameInts(got, tc.want) {
				t.Fatalf("phyrexian_life offered = %v, want %v", got, tc.want)
			}
			dispatchAll(t, g, seat.ID, castMovesFor(moves, loot.InstanceID))
		})
	}
}
