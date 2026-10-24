package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "b20d74a4-0e08-4779-972a-499f1a2be15a",
		Name:            "Mendicant Core, Guidelight",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Static:          []game.StaticAbility{PowerEqualToArtifactsYouControl()},
		Triggered: []game.TriggeredAbility{
			MaxSpeedTrigger(WheneverYouCast(Artifact(), "Mendicant Core, Guidelight — you may pay {1} to copy that artifact spell",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MayPay{
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
var mendicantCorePaid = PayUnlessBranches("pay-unless/mendicant-core-guidelight-copy-spell",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return CopySpell{StackID: ctx.Trigger().Event.CardID, Controller: ctx.Controller()}.Apply(ctx)
	}, nil)
