package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Living Terrain — Enchantment — Aura {2}{G}{G} (Eighth Edition, S82 older sets
// batch 617, #5607):
//
//	"Enchant land
//	 Enchanted land is a 5/6 green Treefolk creature that's still a land."
//
// The Zendikon animation (layer 4 type and subtype, layer 5 colour, layer 7b
// base power and toughness) with a Treefolk and 5/6.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffa40f10-4efb-48d8-b16e-8a5711fcfad1",
		Name:         "Living Terrain",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       effects.S82OldB100AnimateAttachedLand("G", "Treefolk", 5, 6),
	})
}
