package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Descendant of Masumaro — Creature — Human Monk {2}{G}, 1/1 (Saviors of Kamigawa, S82
// OLD b593):
//
//	"At the beginning of your upkeep, put a +1/+1 counter on this creature for each
//	 card in your hand, then remove a +1/+1 counter from this creature for each card
//	 in target opponent's hand."
//
// One upkeep trigger with a target opponent. The removal reads that hand as the
// ability resolves and runs after the counters land (a registered tail, so a counter
// doubler's ordering prompt pauses it correctly); it cannot take more than the creature
// has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5bfb740-8ab2-4160-a8f0-5b70b63c2f3a",
		Name:         "Descendant of Masumaro",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourUpkeep("Descendant of Masumaro — +1/+1 counters for your hand, then remove one for each card in target opponent's hand", effects.DescendantOfMasumaroUpkeep),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
