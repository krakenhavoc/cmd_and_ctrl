package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cragganwick Cremator — Creature — Giant Shaman {2}{R}{R}, 5/4 (Shadowmoor,
// S82 OLD b530, #5591):
//
//	"When this creature enters, discard a card at random. If you discard a
//	 creature card this way, this creature deals damage equal to that card's
//	 power to target player or planeswalker."
//
// The target is chosen as the trigger goes on the stack. As it resolves the
// random discard is made and the card read back off the discard event it just
// logged (Reckless Handling's shape); only a creature card deals damage, equal
// to its power, from the Cremator or from its last-known information if it has
// left (CR 608.2h). An empty hand discards nothing and deals nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8ea46e2-84ca-4dcb-b675-a2374bc967bc",
		Name:         "Cragganwick Cremator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Cragganwick Cremator — discard a card at random; if it's a creature card, damage equal to its power to target player or planeswalker",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					mark := effects.B25LastEventSeq(g)
					if err := g.DiscardRandomForEffect(item.Controller, 1); err != nil {
						return err
					}
					discarded, ok := effects.B25DiscardedByAfter(g, item.Controller, mark)
					if !ok || !discarded.IsCreature() || discarded.Power <= 0 {
						return nil
					}
					target, ok := effects.FirstLegalTarget(ctx)
					if !ok {
						return nil
					}
					ref, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					return effects.DealDamage{SourceObject: &ref, Target: target.ID, Amount: discarded.Power}.Apply(ctx)
				}),
				effects.TargetPlayerOrPlaneswalker()),
		},
	})
}
