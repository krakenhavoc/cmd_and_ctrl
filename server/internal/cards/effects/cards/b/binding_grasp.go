package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Binding Grasp — Enchantment — Aura {3}{U} (Ice Age, S82 OLD b828, #5636):
//
//	"Enchant creature
//	 At the beginning of your upkeep, sacrifice this Aura unless you pay {1}{U}.
//	 You control enchanted creature.
//	 Enchanted creature gets +0/+1."
//
// Control Magic's layer-2 control change, PumpAttached, and Melancholy's
// upkeep tax (UpkeepSacrificeUnlessYouPay).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1ebdc091-16d5-4803-821a-ac8c4696b5bc",
		Name:         "Binding Grasp",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.ControlAttachedBySource(), effects.PumpAttached(0, 1)},
		Triggered:    []game.TriggeredAbility{effects.UpkeepSacrificeUnlessYouPay("Binding Grasp", "{1}{U}")},
	})
}
