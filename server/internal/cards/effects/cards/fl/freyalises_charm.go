package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freyalise's Charm — Enchantment {G}{G} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Whenever an opponent casts a black spell, you may pay {G}{G}. If you do, you
//	 draw a card.
//	 {G}{G}: Return this enchantment to its owner's hand."
//
// Leshrac's Sigil's shape: the trigger is on the cast of an opponent's black
// spell (the colour read off the spell on the stack), resolving it asks the
// Charm's controller to pay {G}{G} (MayPay), and the paid branch draws. The
// second ability returns the Charm as an effect (not a cost), so it can be
// answered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e0d638d2-2c5b-415c-95e7-74bf8d88794f",
		Name:         "Freyalise's Charm",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.AnOpponentCast(effects.OfColor("B")),
				"Freyalise's Charm — you may pay {G}{G}; if you do, you draw a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{G}{G}",
						Question: "Freyalise's Charm — pay {G}{G} to draw a card?",
						Then:     freyalisesCharmPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{G}: Return this enchantment to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{G}{G}"),
			Effect:  effects.ReturnThisPermanentToOwnersHand,
		}},
	})
}

// freyalisesCharmPaid is the paid branch: the Charm's controller draws. An
// on-disk key.
var freyalisesCharmPaid = effects.PayUnlessBranches("pay-unless/freyalises-charm-draw",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		return effects.DrawCards{Player: r.Chooser, N: 1}.Apply(ctx)
	}, nil)
