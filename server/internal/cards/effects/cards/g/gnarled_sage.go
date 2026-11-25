package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gnarled Sage — Creature — Treefolk Druid {3}{G}{G}, 4/4 (Core Set 2021,
// #5511):
//
//	"Reach
//	 As long as you've drawn two or more cards this turn, this creature gets
//	 +0/+2 and has vigilance."
//
// One printed ability, two layered effects over the same condition: the
// toughness bonus in layer 7c and the keyword in layer 6, both read from the
// controller's draw tally for the turn.
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7C_Modify,
		Label:             "As long as you've drawn two or more cards this turn, this creature gets +0/+2.",
		DependsOnHandSize: true,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			return effects.S82BroB07DrewTwoOrMore(target, g, source)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.Toughness += 2
		},
	}
	effects.Register(effects.Spec{
		OracleID:        "c36d7c2b-204f-432e-b8e2-ffb3ffb49231",
		Name:            "Gnarled Sage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Static:          []game.StaticAbility{pump, effects.S82BroB07SelfKeywordWhileDrawnTwo("vigilance")},
	})
}
