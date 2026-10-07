package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// phyrexian_grant_test.go — ADR 0131 (#2531), PR 1: K'rrik, Son of
// Yawgmoth makes every {B} in its controller's costs payable with 2
// life. The enumerator must offer the same life counts it offers for a
// printed {B/P} (#1677, #917), because the engine accepts them: a bot
// behind K'rrik holding one Swamp is offered a {1}{B}{B} spell, and
// every move offered dispatches.

const oracleKrrik = "cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd"

// krrikOn seats the real catalog K'rrik under `p`.
func krrikOn(g *game.Game, p *game.Player) uuid.UUID {
	return battlefieldCard(g, p, game.Card{
		Name: "K'rrik, Son of Yawgmoth", TypeLine: "Legendary Creature — Phyrexian Horror",
		ManaCost: "{4}{B/P}{B/P}{B/P}", OracleID: oracleKrrik, Power: 2, Toughness: 2,
	})
}

// blackSpellTable seats a {1}{B}{B} instant in hand on a main phase
// with `swamps` Swamps at `life`, with or without K'rrik. Dismember's
// shape, written with plain {B}.
func blackSpellTable(t *testing.T, swamps, life int, withKrrik bool) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	seat.Life = life
	for i := 0; i < swamps; i++ {
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
	}
	if withKrrik {
		krrikOn(g, seat)
	}
	card := handCard(seat, game.Card{
		Name: "Black Spell", TypeLine: "Instant", ManaCost: "{1}{B}{B}", Layout: "normal",
	})
	return g, seat, card
}

// The table is dismember's, one symbol family over: with K'rrik the
// counts are the printed Phyrexian ones, without it a short seat has no
// cast at all.
func TestKrrikOffersTheLifeCountsAPrintedPhyrexianSymbolDoes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		swamps int
		life   int
		krrik  bool
		want   []int
	}{
		{"one Swamp at 20 life", 1, 20, true, []int{2}},
		{"two Swamps at 20 life", 2, 20, true, []int{1, 2}},
		{"three Swamps at 20 life", 3, 20, true, []int{0, 2}},
		{"one Swamp at 3 life", 1, 3, true, nil},
		{"two Swamps at 3 life", 2, 3, true, []int{1}},
		{"no Swamps", 0, 20, true, nil},
		// Without K'rrik it is a plain {1}{B}{B}: nothing to buy with life.
		{"one Swamp, no K'rrik", 1, 20, false, nil},
		{"two Swamps, no K'rrik", 2, 20, false, nil},
		{"three Swamps, no K'rrik", 3, 20, false, []int{0}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, seat, card := blackSpellTable(t, tc.swamps, tc.life, tc.krrik)
			moves := legal.EnumerateFor(g, seat.ID)
			// lifeCounts also checks Cost.Life == Cost.PhyrexianLife ==
			// 2 per symbol, and no cost on a mana payment.
			if got := lifeCounts(t, moves, card); !sameInts(got, tc.want) {
				t.Fatalf("phyrexian_life offered = %v, want %v", got, tc.want)
			}
			dispatchAll(t, g, seat.ID, castMovesFor(moves, card))
		})
	}
}

// The offered payment charges what it says.
func TestKrrikLifeCastPaysTheLife(t *testing.T) {
	g, seat, card := blackSpellTable(t, 1, 20, true)
	got := castMovesFor(legal.EnumerateFor(g, seat.ID), card)
	if len(got) != 1 {
		t.Fatalf("offered %d casts, want the one 4-life payment", len(got))
	}
	dispatchOne(t, g, seat.ID, got[0])
	if seat.Life != 16 {
		t.Errorf("life after the cast = %d, want 16", seat.Life)
	}
}

// An opponent's K'rrik grants this seat nothing.
func TestAnOpponentsKrrikOffersNothing(t *testing.T) {
	g, seat, card := blackSpellTable(t, 1, 20, false)
	krrikOn(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)])
	if got := lifeCounts(t, legal.EnumerateFor(g, seat.ID), card); len(got) != 0 {
		t.Fatalf("offered %v under an opponent's K'rrik, want no cast", got)
	}
}

// --- activations ---------------------------------------------------------

// A CR 602 activation costing {1}{B}{B} with two Swamps is offered at the
// cost of 2 life under K'rrik, and not otherwise. Birthing Pod's shape
// (hybrid_phyrexian_test.go) with plain {B}.
func TestKrrikOffersAnActivationsBlackSymbolForLife(t *testing.T) {
	for _, withKrrik := range []bool{true, false} {
		g := newTable(t)
		seat := g.Seats[g.Turn.ActiveSeat]
		clearHand(seat)
		pod := battlefieldCard(g, seat, game.Card{
			Name:     "Black Pod",
			TypeLine: "Artifact",
			ActivatedAbilities: []game.ActivatedAbilityShape{{
				Label: "{1}{B}{B}: Mark",
				Cost:  game.AbilityCost{Mana: "{1}{B}{B}"},
				Effect: func(_ *game.Game, _ *game.StackItem) error {
					return nil
				},
			}},
		})
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
		battlefieldCard(g, seat, basic("Swamp", "Swamp"))
		if withKrrik {
			krrikOn(g, seat)
		}
		advanceTo(t, g, game.StepPrecombatMain)
		moves := legal.EnumerateFor(g, seat.ID)
		got := activationsOf(moves, pod)
		if !withKrrik {
			if len(got) != 0 {
				t.Fatalf("activation offered with no K'rrik and two Swamps for {1}{B}{B}: %v", labels(moves))
			}
			continue
		}
		if len(got) != 1 {
			t.Fatalf("activations of the Pod = %d, want 1: %v", len(got), labels(moves))
		}
		if n := phyrexianLifeOf(t, got[0]); n != 1 {
			t.Errorf("phyrexian_life = %d, want 1 — one {B} bought with life", n)
		}
		if got[0].Cost == nil || got[0].Cost.Life != game.PhyrexianLifePerSymbol ||
			got[0].Cost.PhyrexianLife != game.PhyrexianLifePerSymbol {
			t.Errorf("cost = %+v, want Life = PhyrexianLife = %d", got[0].Cost, game.PhyrexianLifePerSymbol)
		}
		dispatchAll(t, g, seat.ID, got)
	}
}

// --- attack taxes ---------------------------------------------------------

const oracleBlackTax = "test-black-attack-tax"

// blackTaxOn drops a permanent carrying a "{B} for each attacker" tax
// under `p`, wrapping the real catalog's taxes.
func blackTaxOn(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	prev := game.CatalogAttackTaxes
	game.CatalogAttackTaxes = func(key string) []game.AttackTax {
		if key == oracleBlackTax {
			return []game.AttackTax{{
				Label:    "Creatures can't attack you unless their controller pays {B} for each of them.",
				ManaCost: func(game.AttackTaxQuery) string { return "{B}" },
			}}
		}
		return prev(key)
	}
	t.Cleanup(func() { game.CatalogAttackTaxes = prev })
	battlefieldCard(g, p, game.Card{Name: "Black Tax", TypeLine: "Enchantment", OracleID: oracleBlackTax})
}

// With no black mana and no way to make any, a {B} attack tax is offered
// under K'rrik at the price of 2 life (carried in `phyrexian_life`), and
// not offered without it. Every move offered dispatches.
func TestKrrikOffersAnAttackWhoseTaxIsPaidWithLife(t *testing.T) {
	for _, withKrrik := range []bool{true, false} {
		g := newTable(t)
		me := g.Seats[g.Turn.ActiveSeat]
		them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		clearHand(me)
		bear := freshCreature(g, me, "Bear")
		blackTaxOn(t, g, them)
		if withKrrik {
			krrikOn(g, me)
		}
		advanceTo(t, g, game.StepDeclareAttackers)
		moves := legal.EnumerateFor(g, me.ID)
		dispatchAll(t, g, me.ID, moves)

		var taxed *legal.Move
		for i := range moves {
			m := moves[i]
			if m.Kind == legal.KindAttack && m.Source == bear && decodeAttack(t, m).Target == them.ID.String() {
				taxed = &moves[i]
			}
		}
		if !withKrrik {
			if taxed != nil {
				t.Fatalf("a {B} tax with no black mana was offered without K'rrik: %q", taxed.Label)
			}
			continue
		}
		if taxed == nil {
			t.Fatalf("the attack was not offered under K'rrik: %v", labels(moves))
		}
		if got := phyrexianLifeOf(t, *taxed); got != 1 {
			t.Errorf("phyrexian_life = %d, want 1", got)
		}
		if taxed.Cost == nil || taxed.Cost.Life != 2 || taxed.Cost.PhyrexianLife != 2 || taxed.Cost.Mana != "{B}" {
			t.Errorf("cost = %+v, want Mana {B}, Life 2, PhyrexianLife 2", taxed.Cost)
		}
		before := me.Life
		dispatchOne(t, g, me.ID, *taxed)
		if me.Life != before-2 {
			t.Errorf("life after the attack = %d, want %d", me.Life, before-2)
		}
	}
}

// With a Swamp to tap the attack is paid in mana and carries no life.
func TestKrrikAttackTaxPaidInManaWhenTheBoardHasIt(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	bear := freshCreature(g, me, "Bear")
	blackTaxOn(t, g, them)
	krrikOn(g, me)
	battlefieldCard(g, me, basic("Swamp", "Swamp"))
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Kind == legal.KindAttack && m.Source == bear && decodeAttack(t, m).Target == them.ID.String() {
			if n := phyrexianLifeOf(t, m); n != 0 {
				t.Errorf("phyrexian_life = %d with a Swamp to tap, want 0", n)
			}
			if !decodeAttack(t, m).AutoTap {
				t.Error("the Swamp-paid attack does not carry auto_tap")
			}
			return
		}
	}
	t.Fatal("the attack was not offered")
}
