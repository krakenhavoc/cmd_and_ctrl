package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pious Interdiction — Enchantment — Aura {3}{W} (Ixalan, #5532):
//
//	"Enchant creature
//	 When this Aura enters, you gain 2 life.
//	 Enchanted creature can't attack or block."
//
// Pacifism's restriction plus Divine Favor's enters trigger: the Aura's
// controller gains 2 life, and the creature can't attack or block.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6762a65-5c88-4652-ac3d-b9de220634c7",
		Name:         "Pious Interdiction",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pious Interdiction — you gain 2 life", effects.Do(effects.GainLife{Amount: 2})),
		},
	})
}
