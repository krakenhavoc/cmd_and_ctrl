package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voracious Typhon — Creature — Snake Beast {2}{G}{G}, 4/4:
//
//	"Escape—{5}{G}{G}, Exile four other cards from your graveyard.
//	 (You may cast this card from your graveyard for its escape cost.)
//	 This creature escapes with three +1/+1 counters on it."
//
// Escape at its most legible: a vanilla body in the corner and a
// second, bigger body in the graveyard. Nothing about the card is
// conditional on anything except which cost was paid, which makes it
// the reference case for the mechanic — hard-cast it and a 4/4
// arrives, escape it and a 7/7 does, and the difference is three
// counters that hang off the COST rather than off the card.
//
// The part worth stating because flashback trained the opposite
// intuition: escaping does NOT exile the Typhon. It goes to the
// battlefield, and when it dies it is in the graveyard again, ready
// to escape a second time for four more cards. Escape's brake is the
// yard it eats, not a one-shot exile (CR 702.138 has no equivalent
// of flashback's 702.34a clause).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "f62784c0-9c67-4a05-a09b-aabf03a9390f",
		Name:             "Voracious Typhon",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{EscapeWithCounters("{5}{G}{G}", 4, 3)},
	})
}
