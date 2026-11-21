package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnage Interpreter — Creature — Devil Detective {1}{B/R}{B/R}, 3/3
// (Ravnica: Clue Edition):
//
//	"When this creature enters, discard your hand, then investigate four
//	 times.
//	 As long as you have one or fewer cards in hand, this creature gets
//	 +2/+2 and has menace."
//
// The discard is the whole hand even when empty, then four Clues. The
// pump (layer 7c) and menace (layer 6) read the controller's hand on every
// recompute (CR 611.3a).
//
// No simplification.
func init() {
	menace := effects.KeywordGrant(effects.WhileYouHave(effects.BnsB10HandAtMostOne, effects.SelfCreature), "menace")
	menace.Label = "As long as you have one or fewer cards in hand, this creature has menace."
	effects.Register(effects.Spec{
		OracleID:     "4bc79fb3-81bb-42ed-9bbb-93f8cede3094",
		Name:         "Carnage Interpreter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Carnage Interpreter — discard your hand, then investigate four times",
				effects.BnsB10DiscardHandThenInvestigate),
		},
		Static: []game.StaticAbility{
			{
				Label:     "As long as you have one or fewer cards in hand, this creature gets +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: effects.WhileYouHave(effects.BnsB10HandAtMostOne, effects.SelfCreature),
				Apply:     effects.PumpBy(2, 2),
			},
			menace,
		},
	})
}
