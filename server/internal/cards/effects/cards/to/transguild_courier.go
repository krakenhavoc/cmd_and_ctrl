package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transguild Courier — Artifact Creature — Golem {4}, 3/3:
//
//	"Transguild Courier is all colors."
//
// A layer 5 characteristic-changing static on itself that sets all five
// colours (CR 105.2c, 613.1e), the Leyline of the Guildpact's colour
// line aimed at the source alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86f94f8e-f45e-42fc-8ad6-e9bdfaaa997d",
		Name:         "Transguild Courier",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Transguild Courier is all colors.",
			Layer:     game.Layer5Color,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Colors = []string{"W", "U", "B", "R", "G"}
			},
		}},
	})
}
