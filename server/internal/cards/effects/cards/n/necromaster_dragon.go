package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Necromaster Dragon — Creature — Dragon {3}{U}{B}, 4/4 (Dragons of Tarkir):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may pay {2}. If you
//	 do, create a 2/2 black Zombie creature token and each opponent mills two
//	 cards."
//
// Both results hang off the payment (CR 608.2c): nothing happens unless the {2}
// is really paid. Every opponent mills, not only the player that was hit.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b03fc155-2a70-4106-853c-33bdc88d548c",
		Name:            "Necromaster Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Necromaster Dragon — you may pay {2}; if you do, create a 2/2 black Zombie and each opponent mills two cards",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Necromaster Dragon — pay {2} to create a 2/2 black Zombie and have each opponent mill two cards?",
						Then:     necromasterDragonPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var necromasterDragonPaid = effects.PayUnlessBranches("pay-unless/necromaster-dragon-zombie-mill",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if err := (effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1}).Apply(ctx); err != nil {
			return err
		}
		return effects.MillEachPlayerInTurn(ctx, ctx.Opponents(), effects.ConstantMill(2), game.KeyedTail{})
	}, nil)
