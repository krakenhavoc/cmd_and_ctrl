package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tolarian Winds — Instant {1}{U} (EDHREC rank 3152):
//
//	"Discard all the cards in your hand, then draw that many cards."
//
// The one-sided wheel. The hand is discarded whole — every card
// goes, so which one goes first is not a choice the player makes —
// and the draw is the count of what left, read before the draw so
// the new cards are not counted. An empty hand discards nothing and
// draws nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "686cce6e-18ec-45b0-8d2e-74fa353a905e",
		Name:         "Tolarian Winds",
		Completeness: effects.CompletenessFull,
		// The draw waits for a prompt the discard asks (a madness offer;
		// CR 608.2c), and reads the count from the step record (#4163).
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				n, err := effects.DiscardWholeHand(ctx.Game, ctx.Controller())
				ctx.RecordStep(tolarianWindsRecord, nil, n)
				return err
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.DrawCards{N: ctx.StepRecord(tolarianWindsRecord).N()}.Apply(ctx)
			}),
		),
	})
}

// tolarianWindsRecord is the number of cards discarded.
const tolarianWindsRecord = "tolarian-winds/discarded"
