package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clinging Darkness — Enchantment — Aura {1}{B} (Ravnica: City of Guilds,
// #5600):
//
//	"Enchant creature
//	 Enchanted creature gets -4/-1."
//
// Dead Weight's shape with other numbers: the enchant clause is the Aura
// spell's target (CR 303.4a), and the -4/-1 is a layer 7c modification of
// the host (CR 613.4c, pumpEnchantedCreature). A creature left with 0
// toughness dies to the state-based action (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5ad3abf3-1ae5-4c4c-9950-ba7bff0d75ad",
		Name:         "Clinging Darkness",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{pumpEnchantedCreature(-4, -1)},
	})
}
