package cardsw

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "b0394cf2-12a0-4d4f-87e0-fe8937e6faff",
		Name:         "Well of Lost Dreams",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Well of Lost Dreams — pay {X} (up to the life gained) to draw X cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MayPayX{
						Cost:  "{X}",
						Max:   ctx.Trigger().Event.Amount,
						Label: "Well of Lost Dreams",
						Buys:  "draw X cards (X is at most the life you gained: " + strconv.Itoa(ctx.Trigger().Event.Amount) + ")",
						Goal:  effects.XAsHighAsYouCan,
						Unit:  game.PayAmountCards,
						Then:  wellOfLostDreamsPaid,
					}.Apply(ctx)
				}),
		},
	})
}

var wellOfLostDreamsPaid = effects.PayUnlessBranches("pay-unless/well-of-lost-dreams-draw",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.DrawCards{Player: ctx.Controller(), N: r.Number}.Apply(ctx)
	}, nil)
