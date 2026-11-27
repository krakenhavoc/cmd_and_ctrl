package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alloy Golem — Artifact Creature — Golem {6}, 4/4 (Invasion):
//
//	"As this creature enters, choose a color.
//	 This creature is the chosen color. (It's still an artifact.)"
//
// The colour is chosen as it enters (CR 614.12) and stored on the permanent
// (Prism Ring's choice). A layer 5 characteristic-defining-style effect makes it
// exactly that colour (CR 613.1e), replacing its colourless state, and only
// its colour: it is still an artifact. Before an answer exists it stays
// colourless.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "410b7faf-365c-4b73-a64d-bddbe49cfded",
		Name:         "Alloy Golem",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChooseColorAsEnters(game.ColorForBenefit, "Alloy Golem"),
		Static: []game.StaticAbility{{
			Layer:     game.Layer5Color,
			Label:     "This creature is the chosen color.",
			AppliesTo: effects.ThisPermanent,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				if source != nil && source.ChosenColor != "" {
					c.Colors = []string{source.ChosenColor}
				}
			},
		}},
	})
}
