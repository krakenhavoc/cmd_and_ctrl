package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "650bf82e-7f83-470c-bf4a-34281bfe9341",
		Name:         "Battlegrowth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PlusOneCounterOnTheTarget(ctx.Game, item)
		},
	})
}
