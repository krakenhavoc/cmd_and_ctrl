package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lifelink — Enchantment — Aura {W} (Magic 2010, #5585):
//
//	"Enchant creature
//	 Enchanted creature has lifelink. (Damage dealt by the creature also
//	 causes its controller to gain that much life.)"
//
// Timely Ward's shape with lifelink: the Aura grants the keyword to the
// creature it enchants (CR 702.15b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fe8316a-9cff-43e9-a1d5-993a0c9daf3a",
		Name:         "Lifelink",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{GrantToAttached("lifelink")},
	})
}
