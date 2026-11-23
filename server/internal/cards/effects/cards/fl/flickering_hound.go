package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flickering Hound — Creature — Dog {3}{W}, 2/2:
//
//	"Whenever you cast a creature spell, exile up to one other target
//	 creature you control, then return that card to the battlefield
//	 under its owner's control."
//
// Displacer Kitten's shape: an "up to one" target chosen as the
// trigger goes on the stack, then Flicker at resolution. "Other"
// excludes the Hound itself by instance (AnotherTarget). The creature
// returns as a new object under its OWNER's control and re-triggers its
// enters abilities; an empty "up to one" does nothing. The trigger goes
// on the stack above the creature spell that caused it, so the flicker
// resolves first.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "162421d2-8761-437f-bed9-578b61c96f1f",
		Name:         "Flickering Hound",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventCast},
			AppliesTo: effects.YouCast(effects.Creature()),
			TargetsFrom: effects.AnotherTarget(func(other effects.CardPredicate) *game.TargetSpec {
				return effects.TargetCreature("up to one other target creature you control", effects.YouControl(), other).WithCount(0, 1)
			}),
			Key: "Flickering Hound — exile another creature you control, then return it",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return effects.Flicker{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			},
		}},
	})
}
