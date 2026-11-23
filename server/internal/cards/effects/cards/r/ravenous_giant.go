package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Giant — Creature — Giant {2}{R}{R}, 5/5 (Foundations, #4465):
//
//	"At the beginning of your upkeep, this creature deals 1 damage to
//	 you."
//
// DamageController is the footprinted "this deals N damage to you".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46c7a552-4a98-4e1a-8651-8fb0e7153aec",
		Name:         "Ravenous Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Ravenous Giant — this creature deals 1 damage to you", effects.Do(effects.DamageController{N: 1})),
		},
	})
}
