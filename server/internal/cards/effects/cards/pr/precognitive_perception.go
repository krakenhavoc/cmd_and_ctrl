package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Precognitive Perception — Instant {3}{U}{U} (Ravnica Allegiance):
//
//	"Draw three cards.
//	 Addendum — If you cast this spell during your main phase, instead scry
//	 3, then draw three cards."
//
// Cast in the controller's main phase it scries 3 first and draws three once
// the scry is answered (Scry's keyed tail); otherwise it only draws three.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "275d0b90-a8d6-4388-b7db-34615287f532",
		Name:         "Precognitive Perception",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Scry: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
				return effects.Scry{Player: ctx.Controller(), N: 3, ThenTail: ctx.Tail(precognitivePerceptionDraw)}.Apply(ctx)
			}
			return effects.DrawCards{N: 3}.Apply(ctx)
		},
	})
}

// precognitivePerceptionDraw is "then draw three cards" after the scry. An
// on-disk key: never renamed, never reused.
var precognitivePerceptionDraw = effects.CardTail("then/precognitive-perception-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{N: 3}.Apply(ctx)
})
