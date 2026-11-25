package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death's Approach — Enchantment — Aura {B} (Jumpstart, #5510):
//
//	"Enchant creature
//	 Enchanted creature gets -X/-X, where X is the number of creature cards
//	 in its controller's graveyard."
//
// A layer 7c modification re-read on every recompute (the layer cache is
// dropped whenever a card enters or leaves any graveyard), counting the
// creature cards in the graveyard of whoever controls the enchanted creature
// now, not the Aura's controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a1798d3-33fb-4092-a843-45a25f3b4e82",
		Name:         "Death's Approach",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{{
			Label:     "Enchanted creature gets -X/-X, where X is the number of creature cards in its controller's graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, host *game.Card, g *game.Game, _ *game.Card) {
				n := effects.S82OldB100CreatureCardsInControllersGraveyard(g, host)
				c.Power -= n
				c.Toughness -= n
			},
		}},
	})
}
