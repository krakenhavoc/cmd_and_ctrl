package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mantle of Webs — Enchantment — Aura {1}{G} (Magic Origins):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+3 and has reach."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7549206e-22a9-4d56-b3bc-f6f2f5e87e8a",
		Name:         "Mantle of Webs",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 3),
			effects.GrantToAttached("reach"),
		},
	})
}
