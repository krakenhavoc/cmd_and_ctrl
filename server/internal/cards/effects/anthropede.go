package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Anthropede — Creature — Insect {3}{G}, 3/4 (Duskmourn: House of Horror,
// #4466):
//
//	"Reach
//	 When this creature enters, you may discard a card or pay {2}. When you
//	 do, destroy target Room."
//
// Reach is a printed keyword. The enters trigger is Itzquinth's "you may pay.
// When you do" with an either payment: MayPay asks as the trigger resolves
// (CR 603.5), one prompt whose answer names the branch (discard a card or pay
// {2}, b24DiscardACardOrPay), and a branch the chooser can't pay is not
// offered (CR 118.3). Only a payment really made queues the reflexive trigger
// (CR 603.12), which chooses its target, a Room permanent of any controller
// (a Room you control too, the Anthropede's own controller's included), as it
// goes on the stack, so each player may respond; with no Room it is removed
// for want of a target (CR 603.3d), the cost already paid, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "452d98ae-0fa2-4cff-9e18-eecd76ddb20b",
		Name:            "Anthropede",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Anthropede — you may discard a card or pay {2}", func(g *game.Game, item *game.StackItem) error {
				return MayPay{
					Chooser:  item.Controller,
					Pay:      b24DiscardACardOrPay("{2}"),
					Question: "Anthropede — discard a card or pay {2} to destroy target Room?",
					Then:     anthropedePaid,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// anthropedePaid is the "when you do" of the payment, keyed (#4264). An
// on-disk key: never renamed, never reused.
var anthropedePaid = PayUnlessBranches("pay-unless/anthropede-destroy-room",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return WhenYouDo("Anthropede — destroy target Room", anthropedeDestroyRoomBody).Apply(ctx)
	}, nil)

// anthropedeDestroyRoomBody is the reflexive trigger. An on-disk key: never
// renamed, never reused.
var anthropedeDestroyRoomBody = game.ReflexiveBody("anthropede/destroy-room", simpleBody(destroyFirstLegalCardTarget),
	constTargets(func() *game.TargetSpec {
		return TargetPermanent("target Room", HasSubtype("Room"))
	}))
