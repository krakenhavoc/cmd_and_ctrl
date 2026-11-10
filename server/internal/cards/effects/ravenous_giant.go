package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ravenous Giant — Creature — Giant {2}{R}{R}, 5/5 (Foundations, #4465):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to
//	 you."
//
// DamageController is the footprinted "this deals N damage to you".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46c7a552-4a98-4e1a-8651-8fb0e7153aec",
		Name:         "Ravenous Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Ravenous Giant — this creature deals 1 damage to you", Do(DamageController{N: 1})),
		},
	})
}
