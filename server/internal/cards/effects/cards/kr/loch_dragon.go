package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loch Dragon — Creature — Dragon {U/R}{U/R}{U/R}{U/R}, 3/2 (Throne of
// Eldraine, #5516):
//
//	"Flying
//	 Whenever this creature enters or attacks, you may discard a card. If
//	 you do, draw a card."
//
// Keldon Raider's optional discard whose draw follows only from the discard,
// on an enters-or-attacks trigger.
//
// No simplification.
func init() {
	const label = "Loch Dragon — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:        "e40d47cf-6232-41ec-8865-d12a335dfcee",
		Name:            "Loch Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersOrAttacks(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
