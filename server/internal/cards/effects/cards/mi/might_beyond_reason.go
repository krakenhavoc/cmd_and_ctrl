package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Might Beyond Reason — Instant {3}{G} (Shadows over Innistrad, #5543):
//
//	"Put two +1/+1 counters on target creature.
//	 Delirium — Put three +1/+1 counters on that creature instead if there
//	 are four or more card types among cards in your graveyard."
//
// The graveyard is read as the spell resolves (the spell itself is still on
// the stack, so it does not count). The counters go on the target if it is
// still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "563a8e37-b90e-4ea1-9366-8502b403e13e",
		Name:         "Might Beyond Reason",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 2
			if effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
				n = 3
			}
			return effects.AddCounter{Target: t, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
		},
	})
}
