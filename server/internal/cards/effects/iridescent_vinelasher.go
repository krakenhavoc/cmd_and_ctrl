package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Iridescent Vinelasher — Creature — Lizard Assassin {B}, 1/2
// (Bloomburrow, #4467):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token
//	 copy of it.)
//	 Landfall — Whenever a land you control enters, this creature deals 1
//	 damage to target opponent."
//
// Offspring is the shared keyword (offspring.go). Landfall is an ability
// word (CR 207.2c), here Syndicate Liquidators' trigger with Abraded
// Bluffs' "target opponent", chosen as the trigger goes on the stack and
// re-checked on resolution. The offspring token is a copy, so each
// Vinelasher pings for each land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "4f8c4927-46c2-4c18-b5af-c6288dbaff6d",
		Name:          "Iridescent Vinelasher",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Iridescent Vinelasher"),
			Targeting(Landfall("Iridescent Vinelasher — 1 damage to target opponent (landfall)", DamageToFirstTarget(1)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
