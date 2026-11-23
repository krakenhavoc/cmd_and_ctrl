package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Order of the Golden Cricket — Creature — Kithkin Knight {1}{W}, 2/2 (Jumpstart 2022):
//
//	"Whenever this creature attacks, you may pay {W}. If you do, it gains
//	 flying until end of turn."
//
// Estwald Shieldbasher's MayPay shape with flying; declining, or a pool that
// cannot pay, grants nothing, and a Cricket that left and came back is a new
// object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02abe85a-c603-4963-aba4-524cc80847ea",
		Name:         "Order of the Golden Cricket",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Order of the Golden Cricket — you may pay {W}; if you do, it gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{W}",
						Question: "Order of the Golden Cricket — pay {W} to give it flying until end of turn?",
						Then:     orderOfTheGoldenCricketPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// orderOfTheGoldenCricketPaid grants the keyword. An on-disk key: never
// renamed, never reused.
var orderOfTheGoldenCricketPaid = effects.PayUnlessBranches("pay-unless/order-of-the-golden-cricket-flying",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if ctx.Item == nil || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.GrantKeywordUntilEOT{Target: ctx.Source(), Keywords: []string{"flying"},
			Label: "Order of the Golden Cricket — flying"}.Apply(ctx)
	}, nil)
