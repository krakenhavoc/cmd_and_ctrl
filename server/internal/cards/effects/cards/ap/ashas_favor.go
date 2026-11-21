package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Asha's Favor — Enchantment — Aura {2}{W} (Conflux, #5587):
//
//	"Enchant creature
//	 Enchanted creature has flying, first strike, and vigilance."
//
// Battle Mastery's shape with three keywords: one layer-6 grant to the
// enchanted creature (CR 613.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97e7ada6-dee1-4bb3-a42d-38a61d946bd4",
		Name:         "Asha's Favor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.GrantToAttached("flying", "first strike", "vigilance"),
		},
	})
}
