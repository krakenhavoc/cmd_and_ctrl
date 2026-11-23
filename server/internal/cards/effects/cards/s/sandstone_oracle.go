package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sandstone Oracle — Artifact Creature — Sphinx {7}, 4/4 (Commander 2015,
// #5546):
//
//	"Flying
//	 When this creature enters, choose an opponent. If that player has more
//	 cards in hand than you, draw cards equal to the difference."
//
// The enters trigger is Slithermuse's: the controller chooses the opponent
// (not a target) as it resolves, and the hand sizes are compared then
// (slithermuseDraw). A bad choice is a legal one and draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "850e2dbd-b904-41b9-8ffe-0aa562837bde",
		Name:            "Sandstone Oracle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sandstone Oracle — draw the hand-size difference", func(g *game.Game, item *game.StackItem) error {
				return effects.ChoosePlayer{
					Among:    effects.Opponents,
					Question: "Sandstone Oracle — choose an opponent",
					ThenKey:  effects.SlithermuseDraw,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
