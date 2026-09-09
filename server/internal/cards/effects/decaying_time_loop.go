package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Decaying Time Loop — Instant for {3}{R}:
//
//	"Discard all the cards in your hand, then draw that many cards.
//	 Retrace"
//
// A one-sided Windfall at instant speed. The count is taken before
// the draw, so an empty hand draws nothing — which is the whole
// reason to cast it in response to something rather than on your own
// turn with three cards left.
//
// Retrace ("cast from your graveyard by discarding a land") is an
// alternative cast path with an additional cost (S29) and isn't
// modelled.
func init() {
	Register(Spec{
		OracleID: "13e94530-defb-4c9b-9ec7-bb7789ec2630",
		Name:     "Decaying Time Loop",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n, err := discardWholeHand(ctx.Game, item.Controller)
			if err != nil || n == 0 {
				return err
			}
			return DrawCards{Player: item.Controller, N: n}.Apply(ctx)
		},
	})
}
