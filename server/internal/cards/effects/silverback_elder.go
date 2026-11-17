package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silverback Elder — Creature — Ape Shaman {2}{G}{G}{G}, 5/7 (Dominaria
// United):
//
//	"Whenever you cast a creature spell, choose one —
//	 • Destroy target artifact or enchantment.
//	 • Look at the top five cards of your library. You may put a land
//	   card from among them onto the battlefield tapped. Put the rest on
//	   the bottom of your library in a random order.
//	 • You gain 4 life."
//
// One cast trigger with a real "choose one" (Aetheric Amplifier's modal
// shape): the mode, and the target of the first, are announced as the
// trigger goes on the stack (CR 603.3c), so the other two modes put no
// picker up. The second mode is Elvish Rejuvenator's private look, one
// optional land card put onto the battlefield tapped and the rest on the
// bottom in a random order whether or not one was put.
//
// No simplification.
func init() {
	const label = "Silverback Elder — choose one: destroy target artifact or enchantment; or look at the top five cards of your library, you may put a land card from among them onto the battlefield tapped, the rest on the bottom in a random order; or you gain 4 life"
	trig := WheneverYouCast(Creature(), label, func(*game.Game, *game.StackItem) error { return nil })
	trig.Modes = ChooseOne(
		ModeDoing("Destroy target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			DestroyTheModesTarget),
		ModeDoing("Look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return PutFromLibraryOntoBattlefield{
					Player:   item.Controller,
					Cards:    ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 5),
					Match:    Land(),
					Max:      1,
					Optional: true,
					Tapped:   true,
					Label:    "Silverback Elder — you may put a land card from among them onto the battlefield tapped",
					ThenTail: PutRestOnBottomInRandomOrderTail,
				}.Apply(ctx)
			}),
		ModeDoing("You gain 4 life.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Amount: 4}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "b1a0ae56-77aa-43ad-9aa8-dc1ab779e9d9",
		Name:         "Silverback Elder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Lands: 1},
		Triggered:    []game.TriggeredAbility{trig},
	})
}
