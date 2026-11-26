package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unholy Strength — Enchantment — Aura {B} (Magic 2011, #5579):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+1."
//
// A layer-7c bonus (CR 613.4c) on whatever the Aura is attached to, so it
// moves with the Aura and is gone when the Aura is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "090d88a9-7f2d-4bd1-a30a-7c48d05068be",
		Name:         "Unholy Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 1)},
	})
}
