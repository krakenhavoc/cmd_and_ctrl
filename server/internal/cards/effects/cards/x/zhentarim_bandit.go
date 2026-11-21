package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zhentarim Bandit — {1}{B} Creature — Halfling Rogue 2/1
//
//	"Whenever this creature attacks, you may pay 1 life. If you do, create a
//	 Treasure token."
//
// The life is paid as the ability resolves (CR 118.12): a controller who
// can't pay it (CR 119.4) is not asked, and declining makes no Treasure.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "86a7f685-05d0-4167-b5bb-8f6943c23087",
		Name:         "Zhentarim Bandit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Zhentarim Bandit — you may pay 1 life; if you do, create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(1),
						Question: "Zhentarim Bandit — pay 1 life to create a Treasure token?",
						Then:     zhentarimBanditPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// zhentarimBanditPaid is "if you do, create a Treasure token". An on-disk
// key: never renamed, never reused.
var zhentarimBanditPaid = effects.PayUnlessBranches("pay-unless/zhentarim-bandit-treasure",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
	}, nil)
