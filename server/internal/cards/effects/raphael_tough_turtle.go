package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2c72e6d2-bbb0-42de-b465-438d584cbddc",
		Name:         "Raphael, Tough Turtle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverAnotherCreatureEntersUnderYourControl("Raphael, Tough Turtle — 1 damage to target opponent",
					b36DesertDualDamage),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
