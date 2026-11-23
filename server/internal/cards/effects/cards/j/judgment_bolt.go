package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Judgment Bolt — Instant {3}{R} (Final Fantasy, #4462):
//
//	"Judgment Bolt deals 5 damage to target creature and X damage to
//	 that creature's controller, where X is the number of Equipment you
//	 control."
//
// Minion Missile's shape: the creature's controller is read off the
// target as it sits when the spell resolves (CR 608.2h), and the damage to
// them does not depend on the creature surviving. X is counted as the
// spell resolves, after the first 5 damage is dealt, and an X of 0 deals
// nothing. A target that has become illegal counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aad9f9f2-4c6b-48d0-80bd-77b73a9ea5f1",
		Name:         "Judgment Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 5)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 5}).Apply(ctx); err != nil {
				return err
			}
			x := effects.CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") })
			if x <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: victim.Controller, Amount: x}.Apply(ctx)
		},
	})
}
