package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scar — Instant {B/R} (Shadowmoor, #5591):
//
//	"Put a -1/-1 counter on target creature."
//
// Skeleton Ship's counter row (putACounterOnTheTarget) on a spell (CR
// 122.1a): a -1/-1 counter on a target still legal as the spell
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4eafe158-abff-43e9-aeae-596430c5ce21",
		Name:         "Scar",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return putACounterOnTheTarget(game.CounterMinusOne)(ctx.Game, item)
		},
	})
}
