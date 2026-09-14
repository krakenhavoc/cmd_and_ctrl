package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "686cce6e-18ec-45b0-8d2e-74fa353a905e",
		Name:         "Tolarian Winds",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n, err := discardWholeHand(ctx.Game, item.Controller)
			if err != nil {
				return err
			}
			return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
