package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kolaghan Skirmisher — Creature — Human Warrior {1}{B}, 2/2 (Dragons of Tarkir,
// #5551):
//
//	"Dash {2}{B} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"
//
// Dash is the shared alternative cost (Ragavan's, CR 702.109): cast for
// its dash cost it gains haste and returns to its owner's hand at the
// beginning of the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ade4ffa5-6609-44eb-8f1a-45e7203d90e3",
		Name:             "Kolaghan Skirmisher",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{2}{B}")},
	})
}
