package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreams of Laguna — Instant {1}{U} (Final Fantasy, #4462):
//
//	"Surveil 1, then draw a card. (To surveil 1, look at the top card of
//	 your library. You may put it into your graveyard.)
//	 Flashback {3}{U} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// Prismari Charm's first bullet at surveil 1 (CR 701.25a): the draw is
// the surveil's continuation, so it waits for the answer and draws the
// card the surveil left on top (CR 608.2c). Flashback (CR 702.34a) is
// Otherworldly Gaze's pair: the graveyard as a cast source and the
// flashback offer bound to it, exiling the card as it leaves the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "e060bf7a-f266-40b4-9772-e243dcbf965b",
		Name:             "Dreams of Laguna",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerValue, Draws: 1},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{3}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return Surveil{Player: item.Controller, N: 1, ThenTail: ctx.Tail(thenDrawACard)}.Apply(ctx)
		},
	})
}
