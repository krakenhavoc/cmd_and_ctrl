package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raiders' Spoils — Enchantment {3}{B} (Khans of Tarkir):
//
//	"Creatures you control get +1/+0.
//	 Whenever a Warrior you control deals combat damage to a player, you may
//	 pay 1 life. If you do, draw a card."
//
// One trigger per Warrior that hits a player; each lets its own payment
// happen. The Warrior is judged as it was when it dealt the damage.
func init() {
	anthem := effects.B16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsCreature()
	}, 1, 0)
	anthem.Label = "Creatures you control get +1/+0."
	effects.Register(effects.Spec{
		OracleID:     "45fd14b9-d66a-452d-add4-36f512fff9c2",
		Name:         "Raiders' Spoils",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Static:       []game.StaticAbility{anthem},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.CombatDamageToPlayerBy(ev, source.Controller, g) {
					return false
				}
				dealer, ok := g.LookupCardForEffect(ev.Source)
				return ok && dealer.HasSubtype(source.TextSubtype("Warrior"))
			}, "Raiders' Spoils — you may pay 1 life; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.LifePayment(1),
						Question: "Raiders' Spoils — pay 1 life to draw a card?",
						Then:     raidersSpoilsPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var raidersSpoilsPaid = effects.PayUnlessBranches("pay-unless/raiders-spoils-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)
