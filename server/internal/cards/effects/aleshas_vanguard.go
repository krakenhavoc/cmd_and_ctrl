package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alesha's Vanguard — Creature — Orc Warrior {3}{B}, 3/3 (Fate Reforged,
// #5552):
//
//	"Dash {2}{B} (You may cast this spell for its dash cost. If you do,
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
		OracleID:         "be662217-41c2-4caa-89da-ae7f6f817c34",
		Name:             "Alesha's Vanguard",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{2}{B}")},
	})
}
