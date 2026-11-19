package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impeccable Timing — Instant {1}{W}:
//
//	"Impeccable Timing deals 3 damage to target attacking or blocking
//	 creature."
//
// Angelic Page's target clause with a printed three damage; the target is
// re-checked at resolution (CR 608.2b), so a creature that has left combat
// takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "102bbedf-9f84-433a-bd01-e1a4c05fbedb",
		Name:         "Impeccable Timing",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DealDamageToTheTarget(3)(ctx.Game, item)
		},
	})
}
