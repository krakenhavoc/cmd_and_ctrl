package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ajani's Mantra — Enchantment {1}{W} (Magic 2011, #5579):
//
//	"At the beginning of your upkeep, you may gain 1 life."
//
// Clearwater Goblet's optional upkeep trigger at a fixed 1: its
// controller's upkeep only, and the "you may" is asked as it resolves
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f47661e-a107-4d98-8773-0a068a63df49",
		Name:         "Ajani's Mantra",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Ajani's Mantra — gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Ajani's Mantra — gain 1 life?"),
		},
	})
}
