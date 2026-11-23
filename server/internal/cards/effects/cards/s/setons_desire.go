package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seton's Desire — Enchantment — Aura {2}{G} (Dominaria Remastered, #5948):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2.
//	 Threshold — As long as there are seven or more cards in your graveyard,
//	 all creatures able to block enchanted creature do so."
//
// The pump is the shared PumpAttached. The block requirement is Lure's
// (BlockRequirementLure on the enchanted creature) gated on the Aura's
// controller's graveyard, read live: the layer cache drops whenever a card
// enters or leaves a graveyard, so the requirement starts and stops with the
// seventh card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76fe7cfa-c3db-4122-b262-f7c5c90fcdd7",
		Name:         "Seton's Desire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.Labelled(effects.BlockRequirementWhere(game.BlockRequirementLure, func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.AttachedToSource(target, g, source) && effects.FraThresholdMet(g, source.Controller)
			}), "Threshold — All creatures able to block enchanted creature do so while there are seven or more cards in your graveyard."),
		},
	})
}
