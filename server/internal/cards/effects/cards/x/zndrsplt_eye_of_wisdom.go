package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "502849a6-8e65-40f3-b348-a41c4f939768",
		Name:         "Zndrsplt, Eye of Wisdom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Zndrsplt, Eye of Wisdom", "Okaun, Eye of Chaos"),
			effects.AtBeginningOfYourCombat("Zndrsplt, Eye of Wisdom — flip a coin until you lose a flip",
				effects.BbdFlipUntilYouLose("Zndrsplt, Eye of Wisdom — call the coin flip (flip until you lose)")),
			effects.On(game.EventFlipCoin, effects.BbdAPlayerWonACoinFlip, "Zndrsplt, Eye of Wisdom — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
