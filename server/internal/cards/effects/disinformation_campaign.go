package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disinformation Campaign — Enchantment {1}{U}{B}:
//
//	"When this enchantment enters, you draw a card and each opponent
//	 discards a card.
//	 Whenever you surveil, return this enchantment to its owner's hand."
//
// The enters trigger is Arbiter of Woe's discard and draw halves, in the
// printed order. The return is the surveil trigger's, and does nothing once
// the enchantment has left the battlefield or become a new object
// (returnThisPermanentToOwnersHand, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1dd336b0-9b13-4180-b80a-e8ba5d63d04f",
		Name:         "Disinformation Campaign",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Disinformation Campaign — you draw a card and each opponent discards a card",
				Do(
					DrawCards{N: 1},
					Instruction(func(ctx *Context) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
				)),
			cmdr2B27WheneverYouSurveil("Disinformation Campaign — return it to its owner's hand",
				returnThisPermanentToOwnersHand),
		},
	})
}
