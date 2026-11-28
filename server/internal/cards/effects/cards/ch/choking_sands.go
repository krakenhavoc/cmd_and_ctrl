package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Choking Sands — Sorcery {1}{B}{B} (Mirage):
//
//	"Destroy target non-Swamp land. If that land was nonbasic, Choking Sands
//	 deals 2 damage to the land's controller."
//
// The land's controller and whether it is nonbasic are read before it is
// destroyed (CR 608.2h); the damage does not depend on the land actually
// being destroyed (a replacement could stop that), only on its having been
// nonbasic. A non-Swamp is judged by effective land types, so a land that has
// become a Swamp is not a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3bcb7436-f7d9-4bb0-81be-7a3db9d8329f",
		Name:         "Choking Sands",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target non-Swamp land", effects.Land(), effects.Not(effects.Subtype("Swamp"))),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			controller, nonbasic := card.Controller, effects.B10NonbasicLand()(ctx.Game, t.ID, card)
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			if !nonbasic {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
