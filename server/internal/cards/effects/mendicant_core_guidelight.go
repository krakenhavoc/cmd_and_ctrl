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
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of artifacts you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := 0
				for _, p := range g.Battlefield.Cards {
					if p.Controller == source.Controller && p.IsArtifact() {
						n++
					}
				}
				c.Power = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			MaxSpeedTrigger(WheneverYouCast(Artifact(), "Mendicant Core, Guidelight — you may pay {1} to copy that artifact spell",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					spell := ctx.Trigger().Event.CardID
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Mendicant Core, Guidelight — pay {1} to copy that artifact spell?",
						OnPay: func(ctx *Context) error {
							return CopySpell{StackID: spell, Controller: ctx.Controller()}.Apply(ctx)
						},
					}.Apply(ctx)
				})),
		},
	})
}
