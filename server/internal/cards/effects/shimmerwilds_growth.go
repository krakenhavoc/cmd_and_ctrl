package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shimmerwilds Growth — Enchantment — Aura {1}{G} (Lorwyn Eclipsed, #4458):
//
//	"Enchant land
//	 As this Aura enters, choose a color.
//	 Enchanted land is the chosen color.
//	 Whenever enchanted land is tapped for mana, its controller adds an
//	 additional one mana of the chosen color."
//
// The color is the Aura's own (Card.ChosenColor). The land's color is a
// layer 5 change; the mana is a mana trigger (the ruling: a mana ability,
// no stack). With no color chosen it adds nothing.
func init() {
	Register(Spec{
		OracleID:     "c63402b2-1f8c-4004-82d8-48ea26bfe3fc",
		Name:         "Shimmerwilds Growth",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		AsEnters:     ChooseColorAsEnters(game.ColorForMana, "Shimmerwilds Growth"),
		Static: []game.StaticAbility{{
			Label:                 "Enchanted land is the chosen color.",
			Layer:                 game.Layer5Color,
			ContinuesAfterRemoval: true,
			AppliesTo:             AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if source.ChosenColor != "" {
					c.Colors = []string{source.ChosenColor}
				}
			},
		}},
		ManaTriggers: []game.ManaTrigger{
			WheneverAttachedTapsForManaFunc(
				"Shimmerwilds Growth — add an additional one mana of the chosen color",
				AddsOneManaOfTheChosenColor()),
		},
	})
}
