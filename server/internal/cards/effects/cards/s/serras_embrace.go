package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Serra's Embrace — Enchantment — Aura {2}{W}{W} (Foundations Jumpstart):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and has flying and vigilance."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6d6ba936-4a15-4c40-aaa6-71605fb732d1",
		Name:         "Serra's Embrace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("flying"),
			effects.GrantToAttached("vigilance"),
		},
	})
}
