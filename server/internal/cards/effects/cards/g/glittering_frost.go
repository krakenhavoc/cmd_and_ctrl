package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glittering Frost — Snow Enchantment — Aura {2}{G} (Kaldheim, #5505):
//
//	"Enchant land
//	 Enchanted land is snow.
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of any color."
//
// Fertile Ground's triggered mana ability plus a layer-4 supertype
// change (CR 205.4, CR 613.1d): the land gains Snow in addition to its
// other types, so {S} can be paid from it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2013df0b-9e8c-4dbf-87b2-0dec983266fa",
		Name:         "Glittering Frost",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Static: []game.StaticAbility{{
			Label:     "Enchanted land is snow.",
			Layer:     game.Layer4Type,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, s := range c.Supertypes {
					if s == "Snow" {
						return
					}
				}
				c.Supertypes = append(append([]string(nil), c.Supertypes...), "Snow")
			},
		}},
		ManaTriggers: []game.ManaTrigger{
			effects.WheneverAttachedTapsForMana("Glittering Frost — add an additional one mana of any color", "{W|U|B|R|G}"),
		},
	})
}
