package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nimana Skitter-Sneak — Creature — Human Rogue {3}{B}, 3/4 (Zendikar
// Rising, #5508):
//
//	"As long as an opponent has eight or more cards in their graveyard, this
//	 creature gets +1/+0 and has menace. (It can't be blocked except by two
//	 or more creatures.)"
//
// Two statics on the same live condition (Expedition Lookout's read of any
// opponent's graveyard): a layer 7c +1/+0 and a layer 6 menace, both on the
// Sneak alone.
//
// No simplification.
func init() {
	cond := func(g *game.Game, source *game.Card) bool {
		return effects.S82OpponentHasGraveyardCards(g, source.Controller, 8)
	}
	effects.Register(effects.Spec{
		OracleID:     "d3168d01-12ea-4a93-9519-82a3544b5bf6",
		Name:         "Nimana Skitter-Sneak",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "As long as an opponent has eight or more cards in their graveyard, this creature gets +1/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && cond(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.Power++ },
		}, {
			Label: "As long as an opponent has eight or more cards in their graveyard, this creature has menace.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && cond(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "menace")
			},
		}},
	})
}
