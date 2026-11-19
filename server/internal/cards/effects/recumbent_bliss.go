package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "aaea9b83-3db5-4cdb-8e45-5c00a102c3ba",
		Name:         "Recumbent Bliss",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{RestrictAttached(game.CantAttackOrBlock)},
		Triggered: []game.TriggeredAbility{
			Optional(AtYourUpkeep("Recumbent Bliss — you may gain 1 life", Do(GainLife{Amount: 1})),
				"Recumbent Bliss — gain 1 life?"),
		},
	})
}
