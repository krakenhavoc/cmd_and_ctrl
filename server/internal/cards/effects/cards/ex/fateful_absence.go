package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fateful Absence — Instant {1}{W}:
//
//	"Destroy target creature or planeswalker. Its controller
//	 investigates. (They create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// Pongify's shape: read the controller before the destroy, then hand
// the victim the compensation. Investigating is creating a Clue, and
// the engine has no "whenever you investigate" watcher to tell the two
// apart, so the Clue is the whole instruction (Tireless Tracker's
// shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35bba442-1aec-4d33-b502-4c580d61644b",
		Name:         "Fateful Absence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, t.ID)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.Investigate{Player: controller, N: 1}.Apply(ctx)
		},
	})
}
