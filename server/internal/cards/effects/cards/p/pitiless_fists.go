package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pitiless Fists — Enchantment — Aura {3}{G} (Lorwyn Eclipsed, #4458):
//
//	"Enchant creature you control
//	 When this Aura enters, enchanted creature fights up to one target
//	 creature an opponent controls.
//	 Enchanted creature gets +2/+2."
//
// The fight is the enchanted creature against the target, only while both
// are still on the battlefield (CR 701.14b); no target means no fight.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0d46a97a-b1d6-4bf3-b3ba-a5cdfd4c8bf5",
		Name:         "Pitiless Fists",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Pitiless Fists — enchanted creature fights up to one target creature an opponent controls",
					effects.EnchantedCreatureFightsTargetCreature),
				effects.TargetCreature("up to one target creature an opponent controls", effects.OpponentControls()).WithCount(0, 1)),
		},
	})
}
