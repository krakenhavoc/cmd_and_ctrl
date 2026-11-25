package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crosis, the Purger — Legendary Creature — Dragon {3}{U}{B}{R}, 6/6
// (Commander 2017, #5533):
//
//	"Flying
//	 Whenever Crosis deals combat damage to a player, you may pay {2}{B}. If
//	 you do, choose a color, then that player reveals their hand and discards
//	 all cards of that color."
//
// Rith, the Awakener's "you may pay, then choose a color" over the damaged
// player's hand: the payment is asked as the trigger resolves, the color is
// chosen by Crosis's controller only if it was paid, and the damaged player
// then reveals their hand (CR 701.20) and discards every card of that color
// (a multicolored card of that color too, a colorless card never). The
// damaged player is read off the damage event, so a player who left the game
// has no hand to reveal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1ae1dcd-76cd-4714-9df2-1ab0ebd2ca41",
		Name:            "Crosis, the Purger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Crosis — you may pay {2}{B}; if you do, choose a color, then that player reveals their hand and discards all cards of that color",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{B}",
						Question: "Crosis, the Purger — pay {2}{B}?",
						Then:     crosisPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var crosisPaid = effects.PayUnlessBranches("pay-unless/crosis-the-purger-color",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		effects.ChooseColorTail(game.ColorForHarm, ctx.Game, ctx.Controller(), ctx.Source(),
			"Crosis, the Purger — choose a color", ctx.ItemTail(crosisDiscard))
		return nil
	}, nil)

// crosisDiscard reveals the damaged player's hand and discards every card of
// the chosen color. An on-disk key: never renamed, never reused.
var crosisDiscard = effects.CardTail("then/crosis-the-purger-discard-the-color", func(ctx *effects.Context, o game.TailOutcome) error {
	color := o.Tail.Word
	if color == "" || ctx.Item == nil || ctx.Item.Trigger == nil {
		return nil
	}
	victim := ctx.Item.Trigger.Event.Target
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var ids []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasColor(color) {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(victim, ids, ctx.Source(), game.KeyedTail{})
})
