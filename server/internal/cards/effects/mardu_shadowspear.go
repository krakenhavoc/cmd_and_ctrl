package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mardu Shadowspear — Creature — Human Warrior {B}, 1/1 (Fate Reforged,
// #5552):
//
//	"Whenever this creature attacks, each opponent loses 1 life.
//	 Dash {1}{B} (You may cast this spell for its dash cost. If you do,
//	 it gains haste, and it's returned from the battlefield to its
//	 owner's hand at the beginning of the next end step.)"
//
// Vicious Conquistador's attack trigger: each opponent still in the game
// loses 1 life, as a life change and not damage. Dash is the shared Dash
// constructor (freerunning.go): an alternative cost whose entry clause gives
// haste and returns the creature to its owner's hand at the beginning of the
// next end step (CR 702.109a, game/dash.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b6e972ae-cc2d-4c7e-9f49-c0616dc82064",
		Name:             "Mardu Shadowspear",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Dash("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Mardu Shadowspear — each opponent loses 1 life",
				Do(EachOpponentLosesLife{N: 1})),
		},
	})
}
