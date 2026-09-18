package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// hybrid_phyrexian_test.go — the enumerator half of #787. CR 107.4's
// hybrid Phyrexian symbols used to make ParseCost fail, and the #289
// guard in castMovesForCard then dropped the card: a compleated
// planeswalker could not be cast at all, from the board or by a bot.
//
// The cast the enumerator offers is the MANA payment — the life half
// (CR 107.4f) is an announce-time claim the caster makes, not a
// second move. That keeps #695's complaint from spreading: nothing
// here advertises a life payment a life total could not cover,
// because nothing here advertises a life payment.

// TestHybridPhyrexianCastIsEnumerated — with mana for either half of
// the symbol, the cast is offered.
func TestHybridPhyrexianCastIsEnumerated(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	// Ajani, Sleeper Agent's printed cost, on a sorcery so nothing
	// but the cost is under test.
	sage := handCard(active, game.Card{
		Name:     "Compleated Sage",
		TypeLine: "Sorcery",
		ManaCost: "{1}{G}{G/W/P}{W}",
		Layout:   "normal",
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Plains", "Plains"))
	battlefieldCard(g, active, basic("Plains", "Plains"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	seen := false
	for _, m := range moves {
		if m.Source == sage {
			seen = true
		}
	}
	if !seen {
		t.Errorf("a card printing {G/W/P} should be castable off four lands: %v", labels(moves))
	}
}

// TestHybridPhyrexianCastNeedsTheManaForTheSymbol — the offer is still
// affordability-gated: three lands cannot pay a four-symbol cost, and
// the enumerator does not quietly treat the Phyrexian symbol as free
// just because it has a life alternative the move does not claim.
func TestHybridPhyrexianCastNeedsTheManaForTheSymbol(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	sage := handCard(active, game.Card{
		Name:     "Compleated Sage",
		TypeLine: "Sorcery",
		ManaCost: "{1}{G}{G/W/P}{W}",
		Layout:   "normal",
	})
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Forest", "Forest"))
	battlefieldCard(g, active, basic("Plains", "Plains"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	for _, m := range moves {
		if m.Source == sage {
			t.Errorf("offered a cast three lands cannot pay: %q", m.Label)
		}
	}
}
