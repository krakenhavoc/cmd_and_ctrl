package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Darkest Hour — Enchantment {B} (Urza's Saga):
//
//	"All creatures are black."
//
// A layer 5 colour set over every creature on the battlefield, every
// player's (CR 613.1e): each is black and no other colour, for as long
// as Darkest Hour is there. A later colour change still applies over it
// in timestamp order. The colour word follows a text change on Darkest
// Hour, as Eye of Nidhogg's does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5667376e-e59c-4b17-b096-5d92cdfe3db1",
		Name:         "Darkest Hour",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer5Color,
			Label:     "All creatures are black.",
			AppliesTo: b29AllCreatures,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
				c.Colors = []string{source.TextColor("B")}
			},
		}},
	})
}
