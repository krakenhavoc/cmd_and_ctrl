package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1dd336b0-9b13-4180-b80a-e8ba5d63d04f",
		Name:         "Disinformation Campaign",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Disinformation Campaign — you draw a card and each opponent discards a card",
				effects.Do(
					effects.DrawCards{N: 1},
					effects.Instruction(func(ctx *effects.Context) error { return effects.EachOpponentDiscardsOne(ctx.Game, ctx.Item) }),
				)),
			effects.Cmdr2B27WheneverYouSurveil("Disinformation Campaign — return it to its owner's hand",
				effects.ReturnThisPermanentToOwnersHand),
		},
	})
}
