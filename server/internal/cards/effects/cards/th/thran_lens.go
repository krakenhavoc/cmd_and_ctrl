package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thran Lens — Artifact {2}:
//
//	"All permanents are colorless."
//
// A layer-5 colour-setting static over every permanent (CR 613.1e), Thran
// Lens itself included; cards off the battlefield and spells keep their colour.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ac723bf-ce86-4028-a693-0ef498e4c82d",
		Name:         "Thran Lens",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer5Color,
			Label:     "All permanents are colorless.",
			AppliesTo: func(_ *game.Card, _ *game.Game, _ *game.Card) bool { return true },
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Colors = nil
			},
		}},
	})
}
