package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raphael, Tough Turtle — {1}{R} Legendary Creature — Mutant Ninja Turtle 1/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"Alliance — Whenever another creature you control enters, Raphael deals
//	 1 damage to target opponent."
//
// Alliance is an ability word (CR 207.2c): the trigger is "another creature
// you control enters", with HYDRA Assault Robot's 1 damage to the target
// opponent chosen as it goes on the stack, dealt by Raphael.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c72e6d2-bbb0-42de-b465-438d584cbddc",
		Name:         "Raphael, Tough Turtle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverAnotherCreatureEntersUnderYourControl("Raphael, Tough Turtle — 1 damage to target opponent",
					effects.B36DesertDualDamage),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
