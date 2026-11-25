package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hardened-Scale Armor (Global Series Jiang Yanggu & Mu Yanling, #5528):
//
//	Enchant creature
//	Enchanted creature gets +3/+3.
//
// {2}{G} Enchantment — Aura. Moldervine Cloak's pump without dredge.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9eb58db8-7934-485c-8606-fb1a6cc60d42",
		Name:         "Hardened-Scale Armor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
	})
}
