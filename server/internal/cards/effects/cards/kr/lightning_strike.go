package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Strike — Instant {1}{R} (EDHREC rank 2307):
//
//	"Lightning Strike deals 3 damage to any target."
//
// Lightning Bolt at two mana: the same any-target clause and the
// same three damage from the spell as its source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f34b9bc4-7bfe-47fd-ba23-4eeeb46026eb",
		Name:         "Lightning Strike",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove}, // #4734: damages a creature, planeswalker or player
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 3}.Apply(ctx)
		},
	})
}
