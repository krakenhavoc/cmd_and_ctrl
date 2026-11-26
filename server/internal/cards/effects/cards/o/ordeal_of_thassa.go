package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordeal of Thassa — Enchantment — Aura {1}{U} (Theros, S82 OLD b390, #5561):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, put a +1/+1 counter on it. Then
//	 if it has three or more +1/+1 counters on it, sacrifice this Aura.
//	 When you sacrifice this Aura, draw two cards."
//
// Ordeal of Nylea's two abilities: the shared attack trigger, then a
// sacrifice trigger (OrdealSacrificeTrigger) that draws two for the Aura's
// controller.
//
// No simplification.
func init() {
	const name = "Ordeal of Thassa"
	effects.Register(effects.Spec{
		OracleID:     "38e7be61-944e-46f2-be76-fde8b0c19905",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks(name+" — put a +1/+1 counter on it, then sacrifice this Aura if it has three or more",
				effects.OrdealAttackCounterThenSacrifice),
			effects.OrdealSacrificeTrigger(name+" — draw two cards", game.Purpose{Draws: 2}, effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
