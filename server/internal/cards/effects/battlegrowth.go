package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battlegrowth — Instant {G} (Mirrodin, #5606):
//
//	"Put a +1/+1 counter on target creature."
//
// The one-counter body Shortcut to Mushrooms and Bifur share
// (PlusOneCounterOnTheTarget), on an instant. The counter stays after
// the turn ends.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "650bf82e-7f83-470c-bf4a-34281bfe9341",
		Name:         "Battlegrowth",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PlusOneCounterOnTheTarget(ctx.Game, item)
		},
	})
}
