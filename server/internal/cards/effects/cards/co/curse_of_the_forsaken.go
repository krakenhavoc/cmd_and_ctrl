package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Curse of the Forsaken — Enchantment — Aura Curse {2}{W} (Commander 2013, S82
// older sets):
//
//	"Enchant player
//	 Whenever a creature attacks enchanted player, its controller gains 1
//	 life."
//
// One trigger per attacking creature (not batched), controlled by that
// creature's controller, who gains the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e00f8d6b-5bb1-4625-9fa4-114cdc381bd4",
		Name:         "Curse of the Forsaken",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlayer(),
		Triggered: []game.TriggeredAbility{
			effects.OldB383WheneverACreatureAttacksEnchantedPlayer("Curse of the Forsaken — the attacking creature's controller gains 1 life", true,
				effects.Do(effects.GainLife{Amount: 1})),
		},
	})
}
