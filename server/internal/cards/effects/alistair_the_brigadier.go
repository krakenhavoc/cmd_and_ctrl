package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alistair, the Brigadier — Legendary Creature — Human Soldier {1}{G}{W}{U}, 3/3
// (Doctor Who):
//
//	"Whenever you cast a historic spell, create a 1/1 white Soldier creature token.
//	 Whenever Alistair attacks, you may pay {8}. If you do, creatures you control
//	 get +X/+X until end of turn, where X is the number of historic permanents you
//	 control."
//
// Historic is artifact, legendary or Saga (CR 700.6). The attack trigger asks "you
// may pay {8}" as it resolves (Ghostly Pilferer's MayPay); a payment raises the
// pump, which counts the historic permanents as it resolves and gives them to the
// creatures that are there then (CR 611.2c), Alistair himself (legendary) included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2fe8d739-dee5-4220-8102-6fbae7d4fc19",
		Name:         "Alistair, the Brigadier",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(b14Historic(), "Alistair — create a 1/1 white Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1})),
			WheneverThisAttacks("Alistair — you may pay {8}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{8}",
					Question: "Alistair — pay {8} to give creatures you control +X/+X, where X is the number of historic permanents you control?",
					Then:     alistairPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// alistairPaid is the "if you do": +X/+X to the creatures its controller has now.
var alistairPaid = PayUnlessBranches("pay-unless/alistair-the-brigadier-pump",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		x := gapB87HistoricPermanentsYouControl(ctx.Game, ctx.Controller())
		if x <= 0 {
			return nil
		}
		return BoostUntilEOT{
			Match:     And(Creature(), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.Controller == ctx.Controller() }),
			Power:     x,
			Toughness: x,
			Label:     "Alistair, the Brigadier — +X/+X until end of turn",
		}.Apply(ctx)
	}, nil)
