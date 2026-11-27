package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luminous Wake — Enchantment — Aura {2}{W} (Rise of the Eldrazi, S82 OLD
// batch 484, #5581):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks or blocks, you gain 4 life."
//
// Contaminated Bond's trigger shape: an attack event names the enchanted
// creature, and a block event names the blocker once per block declaration
// (Amount <= 1), so a creature that blocks two attackers triggers once
// (CR 509.3a). The life goes to the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "950e345d-455e-4bd4-be8e-bc3381ac3050",
		Name:         "Luminous Wake",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventAttack, game.EventBlock}, effects.OldB484EnchantedAttacksOrBlocks,
				"Luminous Wake — you gain 4 life", effects.Do(effects.GainLife{Amount: 4})),
		},
	})
}
