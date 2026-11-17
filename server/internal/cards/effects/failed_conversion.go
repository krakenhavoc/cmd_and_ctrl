package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Failed Conversion — Enchantment — Aura {4}{B} (March of the Machine,
// S82 MOM batch 13):
//
//	"Enchant creature
//	 Enchanted creature gets -4/-4.
//	 When enchanted creature dies, surveil 2."
//
// A layer 7c -4/-4 on the host (PumpAttached). The death trigger is
// Decomposition's (WhenEnchantedCreatureDies): harvested while the Aura is
// still attached, it surveils for the Aura's controller whoever controlled
// the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0930791-72b0-4d8b-b242-7b4094fed14e",
		Name:         "Failed Conversion",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{PumpAttached(-4, -4)},
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Failed Conversion — surveil 2",
				Do(Surveil{N: 2})),
		},
	})
}
