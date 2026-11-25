package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embereth Skyblazer — Creature — Human Knight {3}{R}, 4/3 (Throne of
// Eldraine, S82 old sets):
//
//	"During your turn, this creature has flying.
//	 Whenever this creature attacks, you may pay {2}{R}. If you do, creatures
//	 you control get +X/+0 until end of turn, where X is the number of
//	 opponents you have."
//
// The flying is Daggersail Aeronaut's layer 6 grant, on only while its
// controller is the active player. The attack trigger asks for the payment as
// it resolves (Copy Catchers' MayPay): paid, every creature its controller
// controls at that moment gets +X/+0, with X the number of opponents still in
// the game (CR 611.2c: creatures that enter later get nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "783f8411-328a-4a00-9610-a058400cbd6d",
		Name:         "Embereth Skyblazer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.ThisPermanent(target, g, source) && effects.IsActivePlayer(g, source.Controller)
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Embereth Skyblazer — you may pay {2}{R}; if you do, creatures you control get +X/+0 until end of turn, where X is the number of opponents you have",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{R}",
						Question: "Embereth Skyblazer — pay {2}{R} to give creatures you control +X/+0 (X = your opponents)?",
						Then:     embrethSkyblazerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// embrethSkyblazerPaid is "creatures you control get +X/+0 until end of turn",
// once {2}{R} is paid. An on-disk key: never renamed, never reused.
var embrethSkyblazerPaid = effects.PayUnlessBranches("pay-unless/embereth-skyblazer-pump",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.BoostUntilEOT{
			Match: effects.And(effects.Creature(), effects.YouControl()),
			Power: len(ctx.Opponents()),
			Label: "Embereth Skyblazer — +X/+0",
		}.Apply(ctx)
	}, nil)
