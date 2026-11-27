package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ordeal of Erebos — Enchantment — Aura {1}{B} (Theros, S82 OLD b390, #5561):
//
//	"Enchant creature
//	 Whenever enchanted creature attacks, put a +1/+1 counter on it. Then
//	 if it has three or more +1/+1 counters on it, sacrifice this Aura.
//	 When you sacrifice this Aura, target player discards two cards."
//
// Ordeal of Purphoros' shape with a player target on the sacrifice trigger,
// chosen as it goes on the stack; the discard is that player's own choice,
// as Mind Rot's is.
//
// No simplification.
func init() {
	const name = "Ordeal of Erebos"
	effects.Register(effects.Spec{
		OracleID:     "e12497ba-2c49-4609-befa-ebba828bedd6",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEnchantedCreatureAttacks(name+" — put a +1/+1 counter on it, then sacrifice this Aura if it has three or more",
				effects.OrdealAttackCounterThenSacrifice),
			effects.Targeting(effects.OrdealSacrificeTrigger(name+" — target player discards two cards", game.Purpose{Discards: 2},
				func(g *game.Game, item *game.StackItem) error {
					return effects.MindRotOnResolve(item, effects.NewContext(g, item))
				}), effects.TargetPlayer("target player")),
		},
	})
}
