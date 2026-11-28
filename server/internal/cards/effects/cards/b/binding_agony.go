package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Binding Agony — Enchantment — Aura {1}{B} (Mirage, S82 OLD batch 811, #5633):
//
//	"Enchant creature
//	 Whenever enchanted creature is dealt damage, this Aura deals that much
//	 damage to that creature's controller."
//
// Soul Link's "is dealt damage" watcher with damage in place of life: "that
// much" is the event's amount after any reduction, and the creature's
// controller is recorded as the trigger goes on the stack, so a creature that
// dies to the damage still costs its controller. The Aura is the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcd41ed1-4a12-416b-8ba0-51a9c0927ddb",
		Name:         "Binding Agony",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB811ReflectDamageToEnchantedCreaturesController("Binding Agony — that much damage to that creature's controller"),
		},
	})
}
