package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cast_timing_test.go — the enumerator half of #1195, in the same
// agreement style cast_gate_test.go uses: every cast the enumerator
// offers is one the engine accepts, and every cast the engine refuses
// is one the enumerator does not offer.
//
// This package used to keep its own copy of CR 307.1 — three lines
// mirroring CastSpell's — and a copy of a rule is a divergence with a
// date on it. Both sides now call game.CastTimingOpenLocked.

const (
	oracleTimingOrrery = "test-legal-timing-orrery"
	oracleTimingTeferi = "test-legal-timing-teferi"
	oracleTimingBear   = "test-legal-timing-bear"
)

// withLegalCastTimings stubs the per-permanent timing declaration,
// chaining to the real catalog so the fixture cannot blank it.
func withLegalCastTimings(t *testing.T, timings map[string][]game.CastTimingRule) {
	t.Helper()
	prev := game.CatalogCastTimings
	game.CatalogCastTimings = func(id string) []game.CastTimingRule {
		if ct, ok := timings[id]; ok {
			return ct
		}
		if prev != nil {
			return prev(id)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogCastTimings = prev })
}

// A creature spell in hand, in the active seat's END step: shut
// without an Orrery, offered with one, and the offer is one the engine
// accepts (dispatchAll casts it).
func TestOrreryPutsACreatureSpellInTheBotsEndStepMoveList(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	bear := handCard(me, game.Card{
		Name: "Timing Bear", TypeLine: "Creature — Bear",
		OracleID: oracleTimingBear, ManaCost: "{1}{G}",
	})
	lands(g, me, "Forest", "Forest", 6)
	advanceTo(t, g, game.StepEnd)

	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), bear)); n != 0 {
		t.Fatalf("a creature spell was offered in an end step with nothing granting flash: %v",
			labels(castMovesFor(legal.EnumerateFor(g, me.ID), bear)))
	}

	withLegalCastTimings(t, map[string][]game.CastTimingRule{
		oracleTimingOrrery: {{
			Timing: game.TimingFlash,
			Label:  "You may cast spells as though they had flash.",
		}},
	})
	battlefieldCard(g, me, game.Card{
		Name: "Timing Orrery", TypeLine: "Artifact", OracleID: oracleTimingOrrery,
	})

	moves := castMovesFor(legal.EnumerateFor(g, me.ID), bear)
	if len(moves) == 0 {
		t.Fatal("the Orrery did not put the creature spell in the move list")
	}
	// #544: and the engine accepts it. This is the half a copy of the
	// rule cannot give you — a bot offered a cast the announce path
	// refuses picks it, is refused, and picks it again.
	dispatchAll(t, g, me.ID, moves[:1])
}

// The inverse, and CR 101.2 with it: an opponent's Teferi takes the
// instant out of a seat's move list even while that seat's own Orrery
// is on the battlefield.
func TestTeferiTakesTheInstantOutOfAnOpponentsMoveList(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(me)
	bolt := handCard(me, game.Card{
		Name: "Timing Bolt", TypeLine: "Instant",
		OracleID: oracleTimingBear, ManaCost: "{R}",
	})
	lands(g, me, "Mountain", "Mountain", 6)
	advanceTo(t, g, game.StepEnd)

	withLegalCastTimings(t, map[string][]game.CastTimingRule{
		oracleTimingOrrery: {{
			Timing: game.TimingFlash,
			Label:  "You may cast spells as though they had flash.",
		}},
		oracleTimingTeferi: {{
			Timing:  game.TimingSorcery,
			Affects: game.TimingAffectsEachOpponent,
			Label:   "Each opponent can cast spells only any time they could cast a sorcery.",
		}},
	})
	battlefieldCard(g, me, game.Card{
		Name: "Timing Orrery", TypeLine: "Artifact", OracleID: oracleTimingOrrery,
	})
	// An instant needs no Orrery; the point is that the Orrery is
	// there and loses anyway.
	if n := len(castMovesFor(legal.EnumerateFor(g, me.ID), bolt)); n == 0 {
		t.Fatal("setup: the instant is not offered at all")
	}

	battlefieldCard(g, them, game.Card{
		Name: "Timing Teferi", TypeLine: "Legendary Planeswalker — Teferi",
		OracleID: oracleTimingTeferi,
	})
	if moves := castMovesFor(legal.EnumerateFor(g, me.ID), bolt); len(moves) != 0 {
		t.Errorf("an opponent's Teferi left %d cast(s) in the move list: %v", len(moves), labels(moves))
	}
}
