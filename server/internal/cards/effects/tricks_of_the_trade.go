package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tricks of the Trade — Enchantment — Aura {3}{U} (Magic 2013, #5568):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and can't be blocked."
//
// Aether Tunnel at +2/+0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fd288d6-445e-4605-b206-aab43e300dde",
		Name:         "Tricks of the Trade",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 0), RestrictAttached(game.CantBeBlocked)},
	})
}
