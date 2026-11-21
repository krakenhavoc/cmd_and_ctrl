package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Slash — Instant {R}:
//
//	"Ferocious — If you control a creature with power 4 or greater,
//	 damage can't be prevented this turn.
//	 Wild Slash deals 2 damage to any target."
//
// Ferocious is checked as the spell resolves, before the damage, and
// gates only the turn grant (ADR 0107 §5): the 2 damage happens either
// way.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f76a0e5b-74c1-4ee8-b502-e8b988086de4",
		Name:         "Wild Slash",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 2)), Answers: game.AnswerRemove}, // #4734: damages a creature, planeswalker or player
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if effects.B24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 4) {
				if err := (effects.DamageCantBePreventedThisTurn{}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DamageToFirstTargetStep(2)(item, ctx)
		},
	})
}
