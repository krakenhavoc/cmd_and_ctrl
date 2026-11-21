package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Turret Ogre — Creature — Ogre Warrior {3}{R}, 4/3 (Ravnica: Clue Edition,
// S82):
//
//	"Reach
//	 When this creature enters, if you control another creature with power 4
//	 or greater, this creature deals 2 damage to each opponent."
//
// Nessian Hornbeetle's intervening "if" (CR 603.4), checked as the creature
// enters and again as the trigger resolves, against layered power; the
// damage is Fear of Burning Alive's each-opponent primitive.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb4a044b-f25d-47ff-b272-754e577ce50a",
		Name:            "Turret Ogre",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you control another creature with power 4 or greater",
				Test: func(ctx *effects.Context) bool {
					return effects.HornbeetleControlsAnotherPowerFour(ctx.Game, ctx.Controller(), ctx.Source())
				},
			}, effects.WhenThisEnters("Turret Ogre — it deals 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2}))),
		},
	})
}
