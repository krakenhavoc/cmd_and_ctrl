package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zndrsplt, Eye of Wisdom — Legendary Creature — Homunculus {4}{U}, 1/4
// (Battlebond, #5529):
//
//	"Partner with Okaun, Eye of Chaos (When this creature enters, target
//	 player may put Okaun into their hand from their library, then
//	 shuffle.)
//	 At the beginning of combat on your turn, flip a coin until you lose a
//	 flip.
//	 Whenever a player wins a coin flip, draw a card."
//
// Okaun's partner, the same two halves with a draw in place of the
// doubling: the shared flip-until-you-lose chain (bbdFlipUntilYouLose) at
// the beginning of combat on its controller's turn, and one card for its
// controller for every coin flip any player wins (bbdAPlayerWonACoinFlip).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "502849a6-8e65-40f3-b348-a41c4f939768",
		Name:         "Zndrsplt, Eye of Wisdom",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			PartnerWith("Zndrsplt, Eye of Wisdom", "Okaun, Eye of Chaos"),
			AtBeginningOfYourCombat("Zndrsplt, Eye of Wisdom — flip a coin until you lose a flip",
				bbdFlipUntilYouLose("Zndrsplt, Eye of Wisdom — call the coin flip (flip until you lose)")),
			On(game.EventFlipCoin, bbdAPlayerWonACoinFlip, "Zndrsplt, Eye of Wisdom — draw a card", Do(DrawCards{N: 1})),
		},
	})
}
