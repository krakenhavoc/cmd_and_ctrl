package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sardian Cliffstomper — Creature — Minotaur Barbarian {1}{R}, 0/4
// (The Brothers' War):
//
//	"As long as it's your turn and you control four or more Mountains,
//	 this creature gets +X/+0, where X is the number of Mountains you
//	 control."
//
// A layer 7c static on itself (Adelbert Steiner's shape) that reads the
// turn and the Mountain count each time the layers are applied: it
// applies only on its controller's turn and only with four or more
// Mountains, and X is the count then (a land that stops being a Mountain
// is not counted).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20f3bc44-c7fa-4d5a-adb9-2dd1d2d66ffe",
		Name:         "Sardian Cliffstomper",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "Sardian Cliffstomper gets +X/+0 on your turn while you control four or more Mountains, X being the number of Mountains you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && effects.IsActivePlayer(g, source.Controller) &&
					effects.B08LandsWithSubtypeControlled(g, source.Controller, source.TextSubtype("Mountain")) >= 4
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += effects.B08LandsWithSubtypeControlled(g, source.Controller, source.TextSubtype("Mountain"))
			},
		}},
	})
}
