package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dead Weight — Enchantment — Aura {B} (The Lost Caverns of Ixalan,
// #4470):
//
//	"Enchant creature
//	 Enchanted creature gets -2/-2."
//
// The enchant clause is the Aura spell's target (CR 303.4a) and is
// re-checked by the state-based action that puts an illegally attached
// Aura into the graveyard (CR 704.5m). The -2/-2 is a layer 7c
// modification of the host while it is a creature (pumpEnchantedCreature,
// Aether Meltdown's); a creature it leaves with 0 toughness dies to the
// state-based action (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1804304-fac1-4b19-a48d-6ade9407972a",
		Name:         "Dead Weight",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static:       []game.StaticAbility{pumpEnchantedCreature(-2, -2)},
	})
}
