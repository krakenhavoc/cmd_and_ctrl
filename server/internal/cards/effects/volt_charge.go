package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Volt Charge — Instant {2}{R}:
//
//	"Volt Charge deals 3 damage to any target. Proliferate. (Choose any
//	 number of permanents and/or players, then give each another counter of
//	 each kind already there.)"
//
// Lightning Bolt's damage, then the controller's proliferate choice
// (Proliferate, #2525). With its only target gone the spell does not
// resolve at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c57907d1-c5a1-45c7-80a8-38fedf86dd19",
		Name:         "Volt Charge",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := damageToFirstTarget(3)(item, ctx); err != nil {
				return err
			}
			return Proliferate{}.Apply(ctx)
		},
	})
}
