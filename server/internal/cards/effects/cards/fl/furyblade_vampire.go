package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furyblade Vampire — Creature — Vampire Berserker {1}{R}, 1/2 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Trample
//	 At the beginning of combat on your turn, you may discard a card. If
//	 you do, this creature gets +3/+0 until end of turn."
//
// Anthropede's "you may pay" with a discard: asked as the trigger resolves
// (CR 603.5), the card chosen by the player, and a player with an empty
// hand can't pay (CR 118.3). Only a card really discarded gives the +3/+0,
// and only to the Vampire that is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8dbfc60-33e9-4068-a0dd-0d0ed5b9b42c",
		Name:            "Furyblade Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Furyblade Vampire — you may discard a card; if you do, +3/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.DiscardPayment(1),
						Question: "Furyblade Vampire — discard a card to get +3/+0 until end of turn?",
						Then:     furybladeVampirePaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// furybladeVampirePaid is "if you do", keyed (#4264). An on-disk key:
// never renamed, never reused.
var furybladeVampirePaid = effects.PayUnlessBranches("pay-unless/furyblade-vampire-pump",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return effects.ThisGetsUntilEndOfTurn(3, 0, "Furyblade Vampire — +3/+0 until end of turn")(ctx.Game, ctx.Item)
	}, nil)
