package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rewards of Diversity — Enchantment {2}{W} (Invasion, #5616):
//
//	"Whenever an opponent casts a multicolored spell, you gain 4 life."
//
// Soldier of the Pantheon's trigger with four life: it watches an opponent
// casting a spell with two or more colours, read through the spell's colours
// as it was cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "064c5a14-c2f2-4b7a-8e0b-01b9995e8f89",
		Name:         "Rewards of Diversity",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.Multicolored()), "Rewards of Diversity — you gain 4 life",
				effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
