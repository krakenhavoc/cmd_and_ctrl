package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Candlelight Vigil — Enchantment — Aura {3}{W} (Guilds of Ravnica, #5525):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+2 and has vigilance."
//
// Sentinel's Eyes' two statics with a larger bonus: a layer 7c pump and a layer
// 6 keyword grant over the enchanted creature, read live (CR 611.3a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7153bf15-53f8-445e-acbb-35b4f4b3d551",
		Name:         "Candlelight Vigil",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 2),
			effects.GrantToAttached("vigilance"),
		},
	})
}
