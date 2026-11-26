package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primal Visitation — Enchantment — Aura {3}{R}{G} (Gatecrash, S82 OLD b407, #5565):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has haste."
//
// Swashbuckling's pump and haste grant at +3/+3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d95e8b1-0ddb-4bd3-b4cc-de55b6f8c403",
		Name:         "Primal Visitation",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("haste"),
		},
	})
}
