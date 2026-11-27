package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Darigaaz, the Igniter — Legendary Creature — Dragon {3}{B}{R}{G}, 6/6 (Invasion, S82 #5616):
//
//	"Flying
//	 Whenever Darigaaz deals combat damage to a player, you may pay {2}{R}. If you
//	 do, choose a color, then that player reveals their hand and Darigaaz deals
//	 damage to the player equal to the number of cards of that color revealed this
//	 way."
//
// Rith, the Awakener's shape. The damaged player is the one the combat damage
// event named; their hand is revealed to the table and the cards of the chosen
// colour in it are counted as the tail runs. The damage is dealt by Darigaaz (a
// source that has since left still deals it, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cbafa7b-4448-4ad4-b4f2-0ab6e7c95b0b",
		Name:            "Darigaaz, the Igniter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Darigaaz, the Igniter — you may pay {2}{R}; if you do, choose a color, that player reveals their hand and Darigaaz deals damage to them equal to the number of cards of that color",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{R}",
						Question: "Darigaaz, the Igniter — pay {2}{R}?",
						Then:     darigaazPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var darigaazPaid = effects.PayUnlessBranches("pay-unless/darigaaz-the-igniter-hand",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, ctx.Controller(), ctx.Source(),
			"Darigaaz, the Igniter — choose a color", ctx.ItemTail(darigaazReveal))
		return nil
	}, nil)

var darigaazReveal = effects.CardTail("then/darigaaz-the-igniter-reveal", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	victim := ctx.Trigger().Event.Target
	if color == "" || victim == uuid.Nil {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	n := 0
	for _, c := range p.Hand.Cards {
		if c.HasColor(color) {
			n++
		}
	}
	if n <= 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: victim, Amount: n}.Apply(ctx)
})
