package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pygmy Kavu — Creature — Kavu {3}{G}, 1/2 (Planeshift, S82 OLD #5615):
//
//	"When this creature enters, draw a card for each black creature your
//	 opponents control."
//
// The count is read as the trigger resolves (CR 608.2c): creatures an opponent
// controls whose colour is black, multicoloured ones included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "17407b73-0dbe-4a71-84ec-3aaca9651f5b",
		Name:         "Pygmy Kavu",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Pygmy Kavu — draw a card for each black creature your opponents control", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				n := 0
				for _, opp := range ctx.Opponents() {
					n += len(effects.PermanentsControlledByMatching(g, opp, effects.And(effects.Creature(), effects.OfColor("B"))))
				}
				return effects.DrawCards{N: n}.Apply(ctx)
			}),
		},
	})
}
