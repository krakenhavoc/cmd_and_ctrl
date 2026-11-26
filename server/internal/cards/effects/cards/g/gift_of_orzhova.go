package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gift of Orzhova — Enchantment — Aura {1}{W/B}{W/B} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has flying and lifelink."
//
// Swashbuckling's shape with two keywords. The hybrid cost is paid with white, black or both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcb6d317-5c55-4973-8e3d-2e98211dc30e",
		Name:         "Gift of Orzhova",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("flying", "lifelink"),
		},
	})
}
