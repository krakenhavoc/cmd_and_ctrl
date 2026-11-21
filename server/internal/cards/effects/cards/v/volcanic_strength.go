package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Strength — Enchantment — Aura {1}{R} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has mountainwalk."
//
// Reach for the Sky's pump and grant at +2/+2 with mountainwalk: a layer
// 7c change and a layer 6 grant on the enchanted creature, both gone when
// the Aura leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c60b7bd-06cc-487f-ba42-b382ac429b12",
		Name:         "Volcanic Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("mountainwalk"),
		},
	})
}
