package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rackling — Artifact Creature — Phyrexian Construct {4}, 2/2 (Nemesis, S82
// OLD, #5618):
//
//	"At the beginning of each opponent's upkeep, this creature deals X damage
//	 to that player, where X is 3 minus the number of cards in their hand."
//
// The upkeep's player is the damaged one, and their hand is counted as the
// trigger resolves; with three or more cards in hand X is not positive and
// nothing is dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41d116d7-3b1e-4777-b3c0-54f8ca6b2785",
		Name:         "Rackling",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachOpponentsUpkeep("Rackling — X damage to that player, X is 3 minus the cards in their hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					who := ctx.Trigger().Event.Actor
					p := g.PlayerByIDForEffect(who)
					if p == nil || p.Hand == nil {
						return nil
					}
					x := 3 - p.Hand.Size()
					if x <= 0 {
						return nil
					}
					return effects.DealDamage{Source: ctx.Source(), Target: who, Amount: x}.Apply(ctx)
				}),
		},
	})
}
