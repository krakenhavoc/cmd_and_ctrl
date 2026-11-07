package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Betrayal at the Vault — Instant {4}{G}{G} (Outlaws of Thunder
// Junction, #4468):
//
//	"Target creature you control deals damage equal to its power to each
//	 of two other target creatures."
//
// Bite Down's two clauses (#764), read positionally: slot 0 is the
// creature that deals the damage, its power read as the spell resolves
// (CR 608.2b) and its deathtouch and lifelink carried, and slot 1 is
// exactly two creatures, any controller's, Distinct from the first
// ("other"). The two recipients are dealt their damage in one instance
// from the biter (DealDamageEachKeyedForEffect, Huntmaster of the Fells'
// call). A biter that left in response deals nothing; a recipient that
// left takes nothing and the other still does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1dce3f0-30c8-488f-bbeb-b0533886c655",
		Name:         "Betrayal at the Vault",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			Distinct(TargetCreature("two other target creatures").WithCount(2, 2)),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			biter, ok := ctx.ClauseTarget(0)
			if !ok || biter.Kind != game.TargetCard {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(biter.ID)
			if !ok {
				return nil
			}
			var victims []uuid.UUID
			for _, t := range ctx.ClauseTargets(1) {
				if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) && !slices.Contains(victims, t.ID) {
					victims = append(victims, t.ID)
				}
			}
			if len(victims) == 0 || c.CurrentPower() <= 0 {
				return nil
			}
			return ctx.Game.DealDamageEachKeyedForEffect(biter.ID, victims, c.CurrentPower(), game.KeyedTail{})
		},
	})
}
