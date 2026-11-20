package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magma Jet — Instant {1}{R} (Fifth Dawn, #5604):
//
//	"Magma Jet deals 2 damage to any target. Scry 2."
//
// Shock's 2 damage, then the scry, which happens whenever the spell
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2f292253-64d3-4cf2-881f-a3eea4fda388",
		Name:         "Magma Jet",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Scry: 2, Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageToFirstTarget(2)(item, ctx); err != nil {
				return err
			}
			return Scry{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
