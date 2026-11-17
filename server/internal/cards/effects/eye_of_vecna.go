package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eye of Vecna — Legendary Artifact {2} (Adventures in the Forgotten
// Realms, #5683):
//
//	"When Eye of Vecna enters, you draw a card and you lose 2 life.
//	 At the beginning of your upkeep, you may pay {2}. If you do, you
//	 draw a card and you lose 2 life."
//
// Both bodies are the same instruction: the draw, then the loss, in one
// resolution. The upkeep payment is decided as the ability resolves
// (MayPay): declining, or having no {2} to pay, does nothing.
//
// Both are triggered abilities, so a Vecna made from an exiled Eye has
// both (The Book of Vile Darkness), and the first triggers when that
// Vecna enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "371a7bbf-5890-46d8-ae57-f020c103b465",
		Name:         "Eye of Vecna",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eye of Vecna — you draw a card and you lose 2 life", func(g *game.Game, item *game.StackItem) error {
				return eyeOfVecnaDrawAndLose(NewContext(g, item))
			}),
			AtYourUpkeep("Eye of Vecna — you may pay {2} to draw a card and lose 2 life", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Cost:     "{2}",
					Question: "Eye of Vecna — pay {2} to draw a card and lose 2 life?",
					Then:     eyeOfVecnaPaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// eyeOfVecnaPaid is the upkeep ability's "If you do", keyed so a table
// waiting on the payer is a restore point.
var eyeOfVecnaPaid = PayUnlessBranches("pay-unless/eye-of-vecna-draw-and-lose",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return eyeOfVecnaDrawAndLose(ctx)
	}, nil)

// eyeOfVecnaDrawAndLose is "you draw a card and you lose 2 life" for the
// ability's controller.
func eyeOfVecnaDrawAndLose(ctx *Context) error {
	you := ctx.Controller()
	if err := (DrawCards{Player: you, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return GainLife{Player: you, Amount: -2}.Apply(ctx)
}
