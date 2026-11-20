package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mardu Scout — Creature — Goblin Scout {R}{R}, 3/1 (Fate Reforged,
// #5552):
//
//	"Dash {1}{R} (You may cast this spell for its dash cost. If you do,
//	 it gains haste, and it's returned from the battlefield to its
//	 owner's hand at the beginning of the next end step.)"
//
// Dash is the shared Dash constructor (freerunning.go): an alternative cost
// whose entry clause gives haste and returns the creature to its owner's
// hand at the beginning of the next end step (CR 702.109a, game/dash.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "c18ab011-c5e0-4174-b623-d89e4d745ce6",
		Name:             "Mardu Scout",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{1}{R}")},
	})
}
