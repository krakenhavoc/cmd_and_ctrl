package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ogw_b1_helpers.go — shared pieces of the Oath of the Gatewatch cards,
// batch 1 (#5545).

// ogwCohortCost is the cost of every cohort ability (CR 207.2c: cohort is
// an ability word with no rules meaning of its own):
//
//	"Cohort — {T}, Tap an untapped Ally you control: …"
//
// The {T} is this creature's own (it waits out summoning sickness, CR
// 302.6). The second half is the tap-others cost (game/tap_others_cost.go):
// not the {T} symbol, so an Ally that arrived this turn may pay it, and it
// does not target. The cohort creature can't pay both halves, because its
// own {T} has already tapped it (CR 118.3), so the clause leaves it out.
func OgwCohortCost() game.AbilityCost {
	return Plus(TapCost(), TapAnotherUntapped("an untapped Ally you control", OfCreatureType("Ally")))
}
