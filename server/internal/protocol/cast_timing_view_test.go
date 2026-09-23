package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cast_timing_view_test.go — #1195. `castable_here` gained its third
// input: the CR 307.1 read CastSpell and the bot enumerator already
// share (game.CastTimingOpenLocked).
//
// The divergence this closes was silent in both directions. A
// flashback SORCERY in a graveyard carried the bit in an opponent's
// end step, where the announce path refuses the cast with
// ErrSorcerySpeedRequired and the bot is offered nothing — and a seat
// with a Vedalken Orrery on the board was told the same card was NOT a
// cast surface on somebody else's turn, when the engine would have
// accepted it. #1024's coherence property never caught either, because
// its fixture happens to sit in a window where the timing answer is
// the same for everybody.

// withCastTimings stubs the per-permanent timing declaration for a set
// of oracle IDs, chaining like every other catalog stub in this
// package.
func withCastTimings(t *testing.T, timings map[string][]game.CastTimingRule) {
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

// timingPermanent puts a permanent declaring `oracle` onto the
// battlefield under `controller`, known to the whole table.
func timingPermanent(g *game.Game, controller *game.Player, name, oracle string) uuid.UUID {
	c := game.NewCard(name, controller.ID)
	c.TypeLine = "Artifact"
	c.OracleID = oracle
	c.Controller = controller.ID
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		if p != nil {
			c.KnownBy[p.ID] = true
		}
	}
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// castSurfaceOf reads the `castable_here` bit one seat is shown for
// one card in one seat's graveyard.
func castSurfaceOf(t *testing.T, g *game.Game, viewer uuid.UUID, pile int, id uuid.UUID) *CardView {
	t.Helper()
	v := ViewOfGameFor(g, viewer.String())
	return cardInSeatZone(t, v.Seats[pile].Graveyard, id)
}

// offeredFromGraveyard reports whether the bot enumerator offers this
// seat a cast of this card out of a graveyard.
func offeredFromGraveyard(g *game.Game, seat, card uuid.UUID) bool {
	for _, m := range legal.EnumerateLocked(g, seat, legal.Options{}) {
		if m.Type != legal.TypeCastSpell || m.Kind != legal.KindCast {
			continue
		}
		if strings.Contains(string(m.Params), card.String()) &&
			strings.Contains(string(m.Params), `"from_zone":"graveyard"`) {
			return true
		}
	}
	return false
}

// THE coherence property, for the input #1024's fixture could not
// vary: whatever the board says about timing, the bit the view stamps
// and the move the enumerator offers are the same answer, for the same
// seat, out of the same zone.
//
// Driven on the ACTIVE seat in its END step, which is the cheapest
// board where the three answers differ: the seat holds priority (the
// enumerator offers nothing at all to a seat that does not, so a
// non-active seat would make every row agree for the wrong reason),
// the stack is empty, and CR 307.1's window is shut because an end
// step is not a main phase.
//
// Three boards, one flashback SORCERY in that seat's own graveyard:
//
//	nothing on the board     — not a cast surface, not a move (CR 307.1)
//	their own Orrery         — a cast surface, and a move
//	plus an opponent's Teferi — neither again (CR 101.2: can't beats can)
func TestViewAndEnumeratorAgreeOnCastTiming(t *testing.T) {
	const (
		oracleFlashback = "test-timing-coherence-flashback"
		oracleOrrery    = "test-timing-coherence-orrery"
		oracleTeferi    = "test-timing-coherence-teferi"
	)
	g := busyTable(t, 1)
	me := g.Seats[g.Turn.ActiveSeat]
	them := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepEnd)

	withCastableZones(t, map[string][]game.ZoneKind{oracleFlashback: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracleFlashback: {{
			Key: "flashback", Label: "Flashback {1}{R}", ManaCost: "{1}{R}",
			FromZone: game.ZoneGraveyard,
		}},
	})
	withCastTimings(t, map[string][]game.CastTimingRule{
		oracleOrrery: {{
			Timing: game.TimingFlash,
			Label:  "You may cast spells as though they had flash.",
		}},
		oracleTeferi: {{
			Timing:  game.TimingSorcery,
			Affects: game.TimingAffectsEachOpponent,
			Label:   "Each opponent can cast spells only any time they could cast a sorcery.",
		}},
	})

	me.Graveyard.Cards = nil
	id := graveyardCard(me, "Coherence Timing Spell", oracleFlashback)
	knownToEveryone(g, me.Graveyard, id)

	check := func(label string, want bool) {
		t.Helper()
		c := castSurfaceOf(t, g, me.ID, me.Seat, id)
		offered := offeredFromGraveyard(g, me.ID, id)
		// The property, said as a property first: the two agree.
		if c.CastableHere != offered {
			t.Errorf("%s: the view says castable_here=%v and the enumerator says %v",
				label, c.CastableHere, offered)
		}
		if c.CastableHere != want {
			t.Errorf("%s: castable_here = %v, want %v", label, c.CastableHere, want)
		}
	}

	check("a sorcery in a graveyard in an end step", false)
	timingPermanent(g, me, "Coherence Orrery", oracleOrrery)
	check("under their own Orrery", true)
	timingPermanent(g, them, "Coherence Teferi", oracleTeferi)
	check("under an opponent's Teferi", false)
}

// `castable_here` is a PER-VIEWER answer wherever the seats disagree
// about timing, and #1037 made the stamps per holder for exactly this
// kind of split. The active seat's own graveyard card is a cast
// surface; the same board tells a non-active seat its own card is not.
func TestCastTimingStampIsPerViewer(t *testing.T) {
	const oracleFlashback = "test-timing-per-viewer-flashback"
	g := busyTable(t, 0)
	active, them := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	if active.ID == them.ID {
		t.Fatal("setup: seat 1 is the active seat")
	}
	withCastableZones(t, map[string][]game.ZoneKind{oracleFlashback: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracleFlashback: {{
			Key: "flashback", Label: "Flashback {1}{R}", ManaCost: "{1}{R}",
			FromZone: game.ZoneGraveyard,
		}},
	})
	active.Graveyard.Cards, them.Graveyard.Cards = nil, nil
	mine := graveyardCard(active, "Active Seat Spell", oracleFlashback)
	theirs := graveyardCard(them, "Other Seat Spell", oracleFlashback)
	knownToEveryone(g, active.Graveyard, mine)
	knownToEveryone(g, them.Graveyard, theirs)

	if c := castSurfaceOf(t, g, active.ID, active.Seat, mine); !c.CastableHere {
		t.Error("the active seat's own flashback sorcery is not a cast surface on its own main phase")
	}
	if c := castSurfaceOf(t, g, them.ID, them.Seat, theirs); c.CastableHere {
		t.Error("a non-active seat's flashback sorcery is a cast surface on somebody else's turn")
	}
}
