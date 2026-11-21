package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unauthorized Exit — Instant {1}{U} (Murders at Karlov Manor, #4469):
//
//	"Return target nonland permanent to its owner's hand. Surveil 1."
//
// The surveil is the bounce's keyed "then" (BounceToHand.ThenTail): a
// commander returned to its owner's hand stops to ask about the command
// zone (CR 903.9), and the surveil waits for that answer. It happens
// whether or not the permanent reached a hand; a target that is illegal on
// resolution makes the whole spell fizzle (CR 608.2b), so no surveil.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d145811-adb2-46a4-b173-bf8152e5de48",
		Name:         "Unauthorized Exit",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			return effects.BounceToHand{Target: t.ID, ThenTail: ctx.Tail(unauthorizedExitSurveil)}.Apply(ctx)
		},
	})
}

var unauthorizedExitSurveil = effects.CardTail("tail/unauthorized-exit-surveil", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.Surveil{N: 1}.Apply(ctx)
})
