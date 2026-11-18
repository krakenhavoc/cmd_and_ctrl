package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Well of Lost Dreams — Artifact {4}:
//
//	"Whenever you gain life, you may pay {X}, where X is less than or
//	 equal to the amount of life you gained. If you do, draw X cards."
//
// X is the payment prompt's number, capped at the life the event gained
// (MayPayX.Max) and read from the trigger event, so it is the amount of
// that one gain. Paying {0} draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0394cf2-12a0-4d4f-87e0-fe8937e6faff",
		Name:         "Well of Lost Dreams",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouGainLife("Well of Lost Dreams — pay {X} (up to the life gained) to draw X cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MayPayX{
						Cost:  "{X}",
						Max:   ctx.Trigger().Event.Amount,
						Label: "Well of Lost Dreams",
						Buys:  "draw X cards (X is at most the life you gained: " + strconv.Itoa(ctx.Trigger().Event.Amount) + ")",
						Goal:  XAsHighAsYouCan,
						Unit:  game.PayAmountCards,
						Then:  wellOfLostDreamsPaid,
					}.Apply(ctx)
				}),
		},
	})
}

var wellOfLostDreamsPaid = PayUnlessBranches("pay-unless/well-of-lost-dreams-draw",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return DrawCards{Player: ctx.Controller(), N: r.Number}.Apply(ctx)
	}, nil)
