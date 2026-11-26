package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Holy Strength — Enchantment — Aura {W} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+2."
//
// A layer-7c bonus (CR 613.4c) on whatever the Aura is attached to.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9357de36-f8be-4f49-b2c8-9fe9eaf82b07",
		Name:         "Holy Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2)},
	})
}
