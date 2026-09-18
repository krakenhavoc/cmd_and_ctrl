package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// x_pay_life_test.go — #957, the pay-life sibling of #810.
//
// `enumeratedXFloor` priced the X in a MANA cost. Toxic Deluge prints
// {2}{B} with no {X} in it and announces X by paying X life
// (`AdditionalCost.PayLifeX`), so the affordable-X search answered 0
// for it and always would: the enumerator offered the cast at X=0 and
// the board was swept for -0/-0. A zero-effect move of exactly #810's
// shape, arriving through a different seam.
//
// The rule is keyed on the COST COMPONENT, not on the card: floor 1
// because Toxic Deluge declares XMatters, ceiling life - 1 because
// CR 119.4 will not let a player pay more life than they have and a
// sweep that kills its own caster is not the one pricing worth
// offering. Any future "as an additional cost, pay X life" is priced
// by the same line.

const oracleToxicDeluge = "afaef788-34d1-460b-b884-9d7ae6ddeb18"

// delugeInHand seeds a castable Toxic Deluge: the card in hand, three
// lands for {2}{B}, and the cursor in the sorcery-speed window.
func delugeInHand(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	clearHand(p)
	id := handCard(p, game.Card{
		Name: "Toxic Deluge", TypeLine: "Sorcery",
		ManaCost: "{2}{B}", OracleID: oracleToxicDeluge,
	})
	battlefieldCard(g, p, basic("Swamp", "Swamp"))
	battlefieldCard(g, p, basic("Island", "Island"))
	battlefieldCard(g, p, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)
	return id
}

// The probe, in the shape the issue reported it: the cast is offered,
// and never again at X=0.
func TestToxicDelugeIsNotOfferedAtXZero(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	deluge := delugeInHand(t, g, active)
	active.Life = 5

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, deluge)
	if len(casts) != 1 {
		t.Fatalf("%d casts of Toxic Deluge off three lands, want 1: %v", len(casts), labels(casts))
	}
	x := xValueOf(t, casts[0])
	if x < 1 {
		t.Errorf("offered at X=%d; a -0/-0 sweep is the move #810 says is not worth offering", x)
	}
	if x > active.Life-1 {
		t.Errorf("offered at X=%d on %d life; CR 119.4 caps the payment and this package stops one short of it",
			x, active.Life)
	}
	dispatchAll(t, g, active.ID, moves)
}

// The ceiling is the seat's life total, one short of it: the largest
// announcement the enumerator makes is the largest sweep the caster
// survives. Four life totals, one rule.
func TestToxicDelugeXIsCappedAtLifeMinusOne(t *testing.T) {
	for _, tc := range []struct {
		life int
		want int
	}{
		{life: 2, want: 1},
		{life: 5, want: 4},
		{life: 13, want: 12},
		// Options.MaxX (20 by default) is the same cap the mana-cost
		// search runs under, and it binds first at a full life total.
		{life: 40, want: 20},
	} {
		g := newTable(t)
		active := g.Seats[g.Turn.ActiveSeat]
		deluge := delugeInHand(t, g, active)
		active.Life = tc.life

		casts := castMovesFor(legal.EnumerateFor(g, active.ID), deluge)
		if len(casts) != 1 {
			t.Fatalf("life %d: %d casts, want 1: %v", tc.life, len(casts), labels(casts))
		}
		if x := xValueOf(t, casts[0]); x != tc.want {
			t.Errorf("life %d: offered X=%d, want %d", tc.life, x, tc.want)
		}
	}
}

// A seat at 1 life has no legal announcement at all: the floor is 1
// (XMatters) and the ceiling is 0, so the cast is not offered —
// the same answer Soothsaying gets with an empty board, and not a
// free X=0 sweep.
func TestToxicDelugeIsNotOfferedAtOneLife(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	deluge := delugeInHand(t, g, active)
	active.Life = 1

	moves := legal.EnumerateFor(g, active.ID)
	if casts := castMovesFor(moves, deluge); len(casts) != 0 {
		t.Errorf("offered %d casts at 1 life: %v", len(casts), labels(casts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// The other half of the rule, and the reason it is keyed on the cost
// component: a card with no pay-X-life cost is priced exactly as it
// was. The Goose Mother is {X}{G}{U} for a 2/2 flier, declares no
// XMatters, and off two lands is still the X=0 cast #810 left alone —
// a life total of 2 does not touch it.
func TestALifeCeilingDoesNotReachACardWithoutAPayLifeCost(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	goose := handCard(active, game.Card{
		Name: "The Goose Mother", TypeLine: "Legendary Creature — Bird Hydra",
		ManaCost: "{X}{G}{U}", Power: 2, Toughness: 2, OracleID: oracleTheGooseMother,
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)
	active.Life = 2

	moves := legal.EnumerateFor(g, active.ID)
	casts := castMovesFor(moves, goose)
	if len(casts) != 1 {
		t.Fatalf("%d casts of The Goose Mother, want 1: %v", len(casts), labels(casts))
	}
	if x := xValueOf(t, casts[0]); x != 0 {
		t.Errorf("X = %d, want 0 — two mana pays for the body and nothing else", x)
	}
	dispatchAll(t, g, active.ID, moves)
}

// The catalog-wide statement, which is what makes this a rule rather
// than one card: EVERY registered spec whose additional cost is a
// pay-X-life is offered at an X the seat can actually pay, never at 0
// when the card declares XMatters, and never above life - 1. The
// engine accepts all of them.
func TestEveryPayLifeXCastIsPricedAndSound(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	// A board to sweep, and mana for the stand-in cost below.
	battlefieldCard(g, active, creature("Mine", "{1}{G}", 2, 2))
	for i := 0; i < 6; i++ {
		battlefieldCard(g, active, basic("Swamp", "Swamp"))
	}
	advanceTo(t, g, game.StepPrecombatMain)
	active.Life = 7

	probed := 0
	for _, spec := range effectsWithPayLifeX() {
		probed++
		// A stand-in printed cost, for the reason the #619 sweep gives:
		// a Spec carries an oracle ID and a name, not a mana cost. The
		// rule under test is the X, so a cost the seat can obviously
		// afford is the probe.
		id := handCard(active, game.Card{
			Name: spec.name, TypeLine: "Sorcery", ManaCost: "{1}{B}", OracleID: spec.oracleID,
		})
		moves := legal.EnumerateFor(g, active.ID)
		casts := castMovesFor(moves, id)
		if len(casts) == 0 {
			t.Errorf("%s: no cast offered on a board that can pay for one", spec.name)
		}
		for _, m := range casts {
			x := xValueOf(t, m)
			if spec.xMatters && x < 1 {
				t.Errorf("%s: %q announced X=%d, and the card says X is the whole of it", spec.name, m.Label, x)
			}
			if x > active.Life-1 {
				t.Errorf("%s: %q announced X=%d on %d life", spec.name, m.Label, x, active.Life)
			}
		}
		dispatchAll(t, g, active.ID, moves)
		removeFromHand(active, id)
	}
	if probed == 0 {
		t.Fatal("no pay-X-life additional cost in the catalog — the sweep is reading the wrong field")
	}
	t.Logf("probed %d pay-X-life costs", probed)
}

type payLifeXSpec struct {
	name     string
	oracleID string
	xMatters bool
}

// effectsWithPayLifeX reads the catalog through the same exported
// hooks the enumerator uses, so the sweep cannot drift from what the
// enumerator sees.
func effectsWithPayLifeX() []payLifeXSpec {
	var out []payLifeXSpec
	for _, spec := range effects.All() {
		c := game.AdditionalCostFor(spec.OracleID)
		if c == nil || !c.PayLifeX {
			continue
		}
		out = append(out, payLifeXSpec{
			name:     spec.Name,
			oracleID: spec.OracleID,
			xMatters: game.XMattersFor(spec.OracleID),
		})
	}
	return out
}
