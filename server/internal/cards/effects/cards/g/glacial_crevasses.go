package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glacial Crevasses — Enchantment {2}{R} (Ice Age, S82 OLD #5636):
//
//	"Sacrifice a snow Mountain: Prevent all combat damage that would be dealt
//	 this turn."
//
// The Fog shape (CR 615.1) with a snow Mountain sacrificed as the cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88e0551a-ada0-41d9-b5c3-39257ce56c3e",
		Name:         "Glacial Crevasses",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a snow Mountain: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent},
			Cost:    effects.SacrificeN(1, "a snow Mountain", effects.Land(), effects.Snow(), effects.HasSubtype("Mountain")),
			Effect:  effects.PreventAllCombatDamageB832("Glacial Crevasses"),
		}},
	})
}
