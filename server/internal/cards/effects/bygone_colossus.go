package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bygone Colossus — Artifact Creature — Robot Giant {9}, 9/9 (EDHREC
// rank 4252):
//
//	"Warp {3} (You may cast this card from your hand for its warp
//	 cost. Exile this creature at the beginning of the next end step,
//	 then you may cast it from exile on a later turn.)"
//
// A colourless 9/9 that can show up on turn three for {3}, swing
// once, and come back later for real. It is in the batch because warp
// is a keyword the engine already implements end to end (#324) —
// which makes the card a plain Spec rather than a mechanic.
//
// The whole card is that one keyword, and the constructor is what
// makes it safe to write. Warp is paid FROM HAND, unlike flashback, so
// the Spec declares no CastableZones: the exile leg and the later
// permission to cast from exile ride the AlternativeCost itself
// (WarpExile), through the same ExilePlayPermission the impulse-exile
// family uses. Writing the price by hand as a bare AlternativeCost
// would ship a 9/9 for {3} that never leaves — not a discount but a
// different, and much better, card.
//
// The delayed exile is "at the beginning of the NEXT end step", not
// "your next end step": a Colossus warped in on an opponent's turn is
// exiled at the end of that turn. That is CR 702.185 and the engine's
// scheduling, not anything declared here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "1bd584d5-4e11-428c-b51e-462e4292b07f",
		Name:             "Bygone Colossus",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Warp("{3}")},
	})
}
