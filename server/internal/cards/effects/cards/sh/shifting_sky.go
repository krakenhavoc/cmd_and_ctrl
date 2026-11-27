package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shifting Sky — Enchantment {2}{U} (Eighth Edition, S82 older sets batch 619, #5607):
//
//	"As this enchantment enters, choose a color.
//	 All nonland permanents are the chosen color."
//
// Painter's Servant's "as enters" choice, but the colour is SET in layer 5
// (CR 613.1e) instead of added: every nonland permanent has exactly the chosen
// colour, whoever controls it. A later colour-changing effect still wins by
// timestamp (CR 613.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4634fd4e-5b38-4cfb-ab73-bf3b14b3bdee",
		Name:         "Shifting Sky",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForBenefit, "Shifting Sky"),
		Static: []game.StaticAbility{{
			Layer: game.Layer5Color,
			Label: "All nonland permanents are the chosen color.",
			AppliesTo: func(card *game.Card, _ *game.Game, source *game.Card) bool {
				return source.ChosenColor != "" && !card.IsLand()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.Colors = []string{source.ChosenColor}
			},
		}},
	})
}
