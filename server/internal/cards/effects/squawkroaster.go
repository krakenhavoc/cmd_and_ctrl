package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squawkroaster — Creature — Elemental {3}{R}, */4 (Lorwyn Eclipsed,
// #4458):
//
//	"Double strike
//	 Vivid — Squawkroaster's power is equal to the number of colors among
//	 permanents you control."
//
// A characteristic-defining ability (CR 604.3, layer 7a) over the controller's
// permanents, Barrowgoyf's shape: the colors are Faeburrow Elder's
// colorsAmongPermanentsYouControl, the Squawkroaster's own red included, so
// it is at least 1/4 on the battlefield. Colorless is not a color and a land
// is colorless (the rulings). The engine carries the printed "*" as a number
// that the ability overrides at every recompute. Vivid is an ability word
// with no rules meaning (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6b79404b-2fae-41fb-994c-d4cd895e8f09",
		Name:            "Squawkroaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{{
			Label:    "Vivid — Squawkroaster's power is equal to the number of colors among permanents you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = len(colorsAmongPermanentsYouControl(g, source.Controller))
			},
		}},
	})
}
