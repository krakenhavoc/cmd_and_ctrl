package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight's Pledge — Enchantment — Aura {1}{W} (Core Set 2019, #5527):
//
//	"Enchant creature
//	Enchanted creature gets +2/+2."
//
// Spectral Steel's Aura half: enchant creature and +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "63604320-68de-415e-b1f9-dd0f85c5d8a3",
		Name:         "Knight's Pledge",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(2, 2)},
	})
}
