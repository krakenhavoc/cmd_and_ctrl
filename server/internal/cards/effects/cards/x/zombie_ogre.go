package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0898fb77-d304-496c-8f2e-1cfe50cc5507",
		Name:         "Zombie Ogre",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.IfVentureCond(effects.CondACreatureDiedThisTurn, effects.AtYourEndStep(effects.VentureLabel("Zombie Ogre"), effects.Venture)),
		},
	})
}
