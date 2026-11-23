package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Recumbent Bliss — Enchantment — Aura {2}{W} (Foundations Jumpstart):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 At the beginning of your upkeep, you may gain 1 life."
//
// The upkeep trigger is the Aura's own, the Aura's controller's upkeep,
// whoever controls the enchanted creature. The gain is a "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aaea9b83-3db5-4cdb-8e45-5c00a102c3ba",
		Name:         "Recumbent Bliss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttackOrBlock)},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Recumbent Bliss — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Recumbent Bliss — gain 1 life?"),
		},
	})
}
