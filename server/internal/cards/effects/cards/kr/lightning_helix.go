package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Helix — "Lightning Helix deals 3 damage to any target
// and you gain 3 life." Composition of two primitives (DealDamage +
// GainLife) in one card file — the canonical shape for S14's
// declarative DSL.
//
// Partial-target handling: if the original target slot is now
// illegal on resolve, the damage half no-ops (DealDamage silently
// returns nil when the target doesn't resolve to a player or
// battlefield card) and the life-gain half fires unconditionally —
// CR 608.2b's "partial illegal still resolves" case lets the
// caster still gain life even when the damage misses.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "800c258a-cfc4-4a54-a667-065ea8dea69e",
		Name:         "Lightning Helix",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove}, // #4734: damages a creature, planeswalker or player
		OnResolve: effects.ResolveSteps(
			// The life gain waits for a prompt the damage asks (Khârn the
			// Betrayer's question, a CR 616 order), CR 608.2c (#4155).
			effects.Instruction(func(ctx *effects.Context) error {
				var target game.TargetRef
				if len(ctx.Item.Targets) > 0 {
					target = ctx.Item.Targets[0]
				}
				return effects.DealDamage{Source: ctx.Source(), Target: target.ID, Amount: 3}.Apply(ctx)
			}),
			effects.GainLife{Amount: 3},
		),
	})
}
