package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suffocating Blast — Instant {1}{U}{U}{R} (Apocalypse, #5614):
//
//	"Counter target spell and Suffocating Blast deals 3 damage to target
//	 creature."
//
// Two clauses, a spell and a creature, each re-judged on its own as the
// Blast resolves (CR 608.2b): when the spell has left the stack the
// creature is still dealt 3, and when the creature has gone the spell is
// still countered. A spell that can't be countered stays on the stack
// and the damage is still dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a9d9433-7b0f-45a3-88aa-9fd7e410c224",
		Name:         "Suffocating Blast",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetSpell("target spell"),
			effects.TargetCreature("target creature"),
		),
		Purpose: effects.ForTargets(effects.DamageToTarget(1, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if s, ok := ctx.ClauseTarget(0); ok {
				if err := (effects.CounterTarget{StackID: s.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
				return effects.DealDamage{Source: ctx.Source(), Target: c.ID, Amount: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
