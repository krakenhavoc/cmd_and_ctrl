package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hylda of the Icy Crown — Legendary Creature — Human Warlock {2}{W}{U},
// 3/4 (Wilds of Eldraine, #4471):
//
//	"Whenever you tap an untapped creature an opponent controls, you may
//	 pay {1}. When you do, choose one —
//	 • Create a 4/4 white and blue Elemental creature token.
//	 • Put a +1/+1 counter on each creature you control.
//	 • Scry 2, then draw a card."
//
// The trigger is Sharae's reader (YouTappedAnOpponentsCreature), once per
// creature tapped. The payment is the ordinary "you may pay" prompt and
// "when you do" is a modal reflexive trigger (Hawkeye's shape, CR 603.12):
// the mode is chosen as it goes on the stack, so it can be answered.
//
// No simplification.
func init() {
	t := effects.On(game.EventTapCard, effects.YouTappedAnOpponentsCreature,
		"Hylda of the Icy Crown — you may pay {1}; when you do, choose one",
		func(g *game.Game, item *game.StackItem) error {
			return effects.MayPay{
				Chooser:  item.Controller,
				Cost:     "{1}",
				Question: "Hylda of the Icy Crown — pay {1}? When you do, choose one",
				Then:     hyldaPaid,
			}.Apply(effects.NewContext(g, item))
		})
	effects.Register(effects.Spec{
		OracleID:     "1733c71b-fd28-48dd-8a3d-41afd4b22f0e",
		Name:         "Hylda of the Icy Crown",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

// hyldaPaid raises the reflexive modal trigger once the {1} is paid, keyed.
var hyldaPaid = effects.PayUnlessBranches("pay-unless/hylda-of-the-icy-crown",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.ReflexiveTrigger{
			Label: "Hylda of the Icy Crown — choose one",
			Body:  hyldaBody,
		}.Apply(ctx)
	}, nil)

var hyldaBody = game.ReflexiveModalBody("hylda-of-the-icy-crown/modes", func(_ uuid.UUID, _ game.EffectParams) *game.ModeSpec {
	return effects.ChooseOne(
		effects.ModeSteps("Create a 4/4 white and blue Elemental creature token.", nil,
			effects.CreateToken{Template: effects.TokenCard("4/4 white and blue Elemental"), N: 1}),
		effects.ModeDoing("Put a +1/+1 counter on each creature you control.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.B11PutCounterOnEachCreatureYouControl(ctx.Game, item)
			}),
		effects.ModeWithPurpose(effects.ModeSteps("Scry 2, then draw a card.", nil, effects.Scry{N: 2}, effects.DrawCards{N: 1}), game.Purpose{Scry: 2, Draws: 1}),
	)
})
