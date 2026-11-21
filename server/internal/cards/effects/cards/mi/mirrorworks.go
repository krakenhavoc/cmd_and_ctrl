package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mirrorworks — Artifact {5}:
//
//	"Whenever another nontoken artifact you control enters, you may pay
//	 {2}. If you do, create a token that's a copy of that artifact."
//
// Hashaton's shape: an entry trigger whose "you may pay {2}. If you do"
// is MayPay, and a CreateTokenCopy of the card the event names (carried
// as the trigger's event, so it is still copied if it has since left,
// CR 707.2). The entering artifact is judged on the battlefield as it
// enters: nontoken, an artifact on the layered view, under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fd9c4ebe-cc9c-427c-8105-adb1703c191b",
		Name:         "Mirrorworks",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, mirrorworksArtifactEntered,
				"Mirrorworks — you may pay {2} to create a token copy of that artifact",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Mirrorworks — pay {2} to create a token that's a copy of that artifact?",
						Then:     mirrorworksPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func mirrorworksArtifactEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
	return ok && !c.IsToken() && c.IsArtifact()
}

// mirrorworksPaid is the "if you do". Keyed: never renamed, never reused.
var mirrorworksPaid = effects.PayUnlessBranches("pay-unless/mirrorworks-copy",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenCopy{
			Controller: ctx.Controller(),
			Copy:       ctx.Trigger().Event.CardID,
			N:          1,
		}.Apply(ctx)
	}, nil)
