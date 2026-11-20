package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lightning Javelin — Sorcery {3}{R} (Magic Origins, #5549):
//
//	"Lightning Javelin deals 3 damage to any target. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)"
//
// Lightning Strike's damage, then scry 1. The scry is last, so nothing
// waits on its prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6f6b062-cd7e-4631-9921-40e054ab835d",
		Name:         "Lightning Javelin",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove, Scry: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
					return err
				}
			}
			return Scry{N: 1}.Apply(ctx)
		},
	})
}
