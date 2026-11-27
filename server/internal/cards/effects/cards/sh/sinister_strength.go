package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sinister Strength — Enchantment — Aura {1}{B} (Planeshift, #5615):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+1 and is black."
//
// A layer-7c +3/+1 and a layer-5 colour that replaces the creature's colours
// with black (Eye of Nidhogg's SetAttachedColors, not "in addition"); both end
// the moment the Aura leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "deb499bd-71d4-4430-8a58-a31f5ab1b239",
		Name:         "Sinister Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 1),
			effects.SetAttachedColors("B"),
		},
	})
}
