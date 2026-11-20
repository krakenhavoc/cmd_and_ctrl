package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e6762a65-5c88-4652-ac3d-b9de220634c7",
		Name:         "Pious Interdiction",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pious Interdiction — you gain 2 life", Do(GainLife{Amount: 2})),
		},
	})
}
