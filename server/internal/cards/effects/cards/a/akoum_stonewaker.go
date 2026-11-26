package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Akoum Stonewaker — Creature — Human Shaman {1}{R}, 2/1 (Battle for
// Zendikar):
//
//	"Landfall — Whenever a land you control enters, you may pay {2}{R}. If
//	 you do, create a 3/1 red Elemental creature token with trample and haste.
//	 Exile that token at the beginning of the next end step."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "36d1a46e-44ac-4a65-a93c-8d6ab5c31cbf",
		Name:         "Akoum Stonewaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Akoum Stonewaker — you may pay {2}{R} to create a 3/1 Elemental with trample and haste (landfall)",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{R}",
						Question: "Akoum Stonewaker — pay {2}{R} to create a 3/1 red Elemental token?",
						Then:     akoumStonewakerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var akoumStonewakerPaid = effects.PayUnlessBranches("pay-unless/akoum-stonewaker-token",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
			Controller: ctx.Controller(),
			Source:     ctx.Item.SourceCardID,
			Groups:     []game.TokenGroup{{Template: ctx.TextToken(effects.TokenCard("3/1 red Elemental with trample and haste")), Count: 1}},
		}, ctx.ItemTail(akoumStonewakerExile))
	}, nil)

var akoumStonewakerExile = effects.SocB8ExileTokensAtNextEndStep("akoum-stonewaker/exile-the-elemental",
	"Akoum Stonewaker — exile the Elemental")
