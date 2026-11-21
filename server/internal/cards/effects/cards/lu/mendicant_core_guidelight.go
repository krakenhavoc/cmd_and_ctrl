package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mendicant Core, Guidelight — Legendary Artifact Creature — Robot {W}{U},
// */3:
//
//	"Mendicant Core's power is equal to the number of artifacts you
//	 control.
//	 Start your engines!
//	 Max speed — Whenever you cast an artifact spell, you may pay {1}. If
//	 you do, copy it. (The copy becomes a token.)"
//
// ADR 0138 (#2122, #2711). The power is a layer 7a characteristic-
// defining ability over the controller's artifacts, itself included.
// The max speed trigger goes on the stack above the artifact spell and
// asks for {1} as it resolves; on payment the spell is copied with the
// engine's CR 707.10 path, which makes a permanent spell's copy a token.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b20d74a4-0e08-4779-972a-499f1a2be15a",
		Name:            "Mendicant Core, Guidelight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Static:          []game.StaticAbility{effects.PowerEqualToArtifactsYouControl()},
		Triggered: []game.TriggeredAbility{
			effects.MaxSpeedTrigger(effects.WheneverYouCast(effects.Artifact(), "Mendicant Core, Guidelight — you may pay {1} to copy that artifact spell",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Mendicant Core, Guidelight — pay {1} to copy that artifact spell?",
						Then:     mendicantCorePaid,
					}.Apply(ctx)
				})),
		},
	})
}

// mendicantCorePaid is the paid half, keyed (#4264): the spell is the asking trigger's event.
var mendicantCorePaid = effects.PayUnlessBranches("pay-unless/mendicant-core-guidelight-copy-spell",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CopySpell{StackID: ctx.Trigger().Event.CardID, Controller: ctx.Controller()}.Apply(ctx)
	}, nil)
