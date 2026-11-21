package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flameshadow Conjuring — Enchantment {3}{R}:
//
//	"Whenever a nontoken creature you control enters, you may pay {R}. If
//	 you do, create a token that's a copy of that creature. That token
//	 gains haste. Exile it at the beginning of the next end step."
//
// Inalla, Archmage Ritualist's shape without the type: the payment is
// offered as the trigger resolves, and the copy is made from the creature
// that entered, read off the trigger's own event, so one that died in
// response is still copied from its last-known values (CR 608.2h). Haste
// is an indefinite grant on the token and the exile is one delayed trigger
// naming it (CR 603.7). Being a token, the copy does not trigger the
// Conjuring again.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "343460ea-7999-4d9e-b425-97fa1621f9d6",
		Name:         "Flameshadow Conjuring",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.NontokenCreatureEnteredUnderYourControl,
				"Flameshadow Conjuring — you may pay {R} to create a hasty token copy of that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{R}",
						Question: "Flameshadow Conjuring — pay {R} to create a token that's a copy of that creature?",
						Then:     flameshadowConjuringPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// flameshadowConjuringPaid is the paid half, keyed (#4264): the creature is
// the asking trigger's event.
var flameshadowConjuringPaid = effects.PayUnlessBranches("pay-unless/flameshadow-conjuring-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenCopiesTogether{
			Controller: ctx.Controller(),
			Copies:     []uuid.UUID{ctx.Trigger().Event.CardID},
			Then:       ctx.ItemTail(flameshadowConjuringRest),
		}.Apply(ctx)
	}, nil)

// flameshadowConjuringRest is "That token gains haste. Exile it at the
// beginning of the next end step." as a keyed tail. An on-disk key: never
// renamed, never reused.
var flameshadowConjuringRest = effects.CardTail("flameshadow-conjuring/haste-and-exile-copy", func(ctx *effects.Context, o game.TailOutcome) error {
	return effects.RprB06HasteAndEndStepTail(ctx, o, "Flameshadow Conjuring", effects.ExileListedCardsBody)
})
