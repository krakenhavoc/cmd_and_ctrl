package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zombie Ogre — Creature — Zombie Ogre {3}{B}{B}, 3/5:
//
//	"At the beginning of your end step, if a creature died this turn,
//	 venture into the dungeon."
//
// An intervening "if" (CR 603.4) on the table-wide count: any creature,
// any controller (TurnTally.CreaturesDied).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0898fb77-d304-496c-8f2e-1cfe50cc5507",
		Name:         "Zombie Ogre",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			IfVentureCond(condACreatureDiedThisTurn, AtYourEndStep(ventureLabel("Zombie Ogre"), Venture)),
		},
	})
}
