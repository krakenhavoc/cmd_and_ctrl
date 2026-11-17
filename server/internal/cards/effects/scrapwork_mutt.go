package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapwork Mutt — Artifact Creature — Dog {2}, 2/1:
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card.
//	 Unearth {1}{R}"
func init() {
	const label = "Scrapwork Mutt — you may discard a card; if you do, draw a card"
	Register(Spec{
		OracleID:     "5906b378-c7a7-44fd-a63b-55a347d97854",
		Name:         "Scrapwork Mutt",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, label, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{Unearth("{1}{R}")},
	})
}
