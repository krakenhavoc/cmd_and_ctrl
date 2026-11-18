package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Immolation — Enchantment — Aura {R} (Innistrad: Midnight Hunt Commander
// reprint, S82):
//
//	"Enchant creature
//	 Enchanted creature gets +2/-2."
//
// Dead Weight's shape: a layer 7c change of the host while it is a
// creature (pumpEnchantedCreature). A host left with 0 or less toughness
// dies to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f40ad89-3767-4837-a078-f2dcfaf368df",
		Name:         "Immolation",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{pumpEnchantedCreature(2, -2)},
	})
}
