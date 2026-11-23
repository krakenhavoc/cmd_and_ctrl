package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tainted Treats — {1}{B}{G} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Destroy target artifact or creature. If its mana value was 4 or
//	 less, create a Food token."
//
// The mana value is the target's as the spell resolves, read before it is
// destroyed (CR 608.2h). The Food does not depend on the destruction: an
// indestructible or regenerated target with a small mana value still makes
// one, since the text says "if its mana value was", not "if you do". A
// target that has become illegal makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60826714-b9fd-4c89-b9f2-1a3c3229686b",
		Name:         "Tainted Treats",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			small := c.ManaValue() <= 4
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if !small {
				return nil
			}
			return effects.CreateToken{Controller: ctx.Controller(), Template: effects.FoodToken(), N: 1}.Apply(ctx)
		},
	})
}
