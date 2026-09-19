package legal_test

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cast_gate_test.go — the enumerator half of #664 and #760, in the
// #499/#618 agreement style: every cast the enumerator OFFERS is one
// the engine accepts, and every cast the engine REFUSES is one the
// enumerator does not offer.
//
// That is the whole reason both read the same two functions. A bot
// offered a banned cast picks it, is refused, and picks it again.

const (
	oracleBurstLightning = "ac2086fe-98ee-4280-9c7c-c5c2d6548a8b"
	oracleRuleOfLaw      = "53e88e64-6f82-4154-a66e-6aeb0154b368"
	oracleUrzasBlast     = "978e0d87-3ff2-4a73-916c-ff0dc0ab2797"
)

// TestKickedAndUnkickedAreBothEnumerated is #664's enumerator half:
// a card with an optional cost is SEVERAL casts, and a bot that was
// only offered the cheap one could never kick anything.
func TestKickedAndUnkickedAreBothEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bolt := handCard(active, game.Card{
		Name: "Burst Lightning", TypeLine: "Instant",
		OracleID: oracleBurstLightning, ManaCost: "{R}",
	})
	// Enough mana for the kicked line ({R} plus {4}).
	lands(g, active, "Mountain", "Mountain", 6)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := castMovesFor(legal.EnumerateFor(g, active.ID), bolt)
	if len(moves) == 0 {
		t.Fatal("Burst Lightning was not offered at all")
	}
	kicked, unkicked := 0, 0
	for _, m := range moves {
		if strings.Contains(m.Label, "Kicker") {
			kicked++
		} else {
			unkicked++
		}
	}
	if kicked == 0 {
		t.Errorf("no kicked line was offered; labels: %v", labels(moves))
	}
	if unkicked == 0 {
		t.Errorf("no unkicked line was offered; labels: %v", labels(moves))
	}
	// #544: every offer is one the engine accepts. dispatchAll casts
	// each in turn against a fresh clone, so the Rule-of-Law-free
	// board cannot refuse the second for the wrong reason.
	dispatchAll(t, g, active.ID, moves[:1])
}

// TestRuleOfLawRemovesTheSecondCastFromTheEnumeration is #760's
// enumerator half, and the agreement that matters: the gate the
// engine refuses with is the gate the enumerator reads.
func TestRuleOfLawRemovesTheSecondCastFromTheEnumeration(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bolt := handCard(active, game.Card{
		Name: "Burst Lightning", TypeLine: "Instant",
		OracleID: oracleBurstLightning, ManaCost: "{R}",
	})
	battlefieldCard(g, active, game.Card{
		Name: "Rule of Law", TypeLine: "Enchantment", OracleID: oracleRuleOfLaw,
	})
	lands(g, active, "Mountain", "Mountain", 6)
	advanceTo(t, g, game.StepPrecombatMain)

	// Nothing cast yet: the bolt is offered.
	if n := len(castMovesFor(legal.EnumerateFor(g, active.ID), bolt)); n == 0 {
		t.Fatal("the first spell of the turn was not offered under Rule of Law")
	}

	// One spell cast this turn: the enumerator offers no cast at all.
	// Set on the tally directly rather than by casting, so the only
	// thing that changed between the two enumerations is the number
	// Rule of Law reads.
	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[active.ID] = game.CastTally{Total: 1}
	})
	for _, m := range legal.EnumerateFor(g, active.ID) {
		if m.Kind == legal.KindCast {
			t.Errorf("a cast was offered under Rule of Law after one spell: %q", m.Label)
		}
	}
}

// TestLegendarySorceryIsNotEnumeratedWithoutALegendary is the
// spell's-own-condition half of the same agreement.
func TestLegendarySorceryIsNotEnumeratedWithoutALegendary(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	blast := handCard(active, game.Card{
		Name: "Urza's Ruinous Blast", TypeLine: "Legendary Sorcery",
		OracleID: oracleUrzasBlast, ManaCost: "{4}{W}",
	})
	lands(g, active, "Plains", "Plains", 8)
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(castMovesFor(legal.EnumerateFor(g, active.ID), blast)); n != 0 {
		t.Errorf("a legendary sorcery was offered with no legendary permanent out")
	}

	battlefieldCard(g, active, game.Card{
		Name: "Legendary Bear", TypeLine: "Legendary Creature — Bear", Power: 2, Toughness: 2,
	})
	moves := castMovesFor(legal.EnumerateFor(g, active.ID), blast)
	if len(moves) == 0 {
		t.Fatal("a legendary sorcery was not offered with a legendary creature out")
	}
	dispatchAll(t, g, active.ID, moves[:1])
}
