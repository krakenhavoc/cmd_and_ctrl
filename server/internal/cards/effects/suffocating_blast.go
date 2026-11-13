package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2a9d9433-7b0f-45a3-88aa-9fd7e410c224",
		Name:         "Suffocating Blast",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetSpell("target spell"),
			TargetCreature("target creature"),
		),
		Purpose: ForTargets(DamageToTarget(1, 3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if s, ok := ctx.ClauseTarget(0); ok {
				if err := (CounterTarget{StackID: s.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if c, ok := ctx.ClauseTarget(1); ok && c.Kind == game.TargetCard {
				return DealDamage{Source: ctx.Source(), Target: c.ID, Amount: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
