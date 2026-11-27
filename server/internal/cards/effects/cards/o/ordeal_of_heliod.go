package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordeal of Heliod — Enchantment — Aura {1}{W} (Theros, S82 OLD b390, #5561):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, put a +1/+1 counter on it. Then
//	 if it has three or more +1/+1 counters on it, sacrifice this Aura.
//	 When you sacrifice this Aura, you gain 10 life."
//
// Ordeal of Nylea's two abilities with a sacrifice trigger that gains the
// Aura's controller 10 life.
//
// No simplification.
func init() {
	const name = "Ordeal of Heliod"
	effects.Register(effects.Spec{
		OracleID:     "a6861402-4819-4d85-8b60-677471f40d7d",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks(name+" — put a +1/+1 counter on it, then sacrifice this Aura if it has three or more",
				effects.OrdealAttackCounterThenSacrifice),
			effects.OrdealSacrificeTrigger(name+" — you gain 10 life", game.Purpose{LifeGain: 10}, effects.Do(effects.GainLife{Amount: 10})),
		},
	})
}
