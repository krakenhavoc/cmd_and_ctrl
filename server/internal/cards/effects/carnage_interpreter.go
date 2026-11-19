package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	menace := KeywordGrant(whileYouHave(bnsB10HandAtMostOne, selfCreature), "menace")
	menace.Label = "As long as you have one or fewer cards in hand, this creature has menace."
	Register(Spec{
		OracleID:     "4bc79fb3-81bb-42ed-9bbb-93f8cede3094",
		Name:         "Carnage Interpreter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Carnage Interpreter — discard your hand, then investigate four times",
				bnsB10DiscardHandThenInvestigate),
		},
		Static: []game.StaticAbility{
			{
				Label:     "As long as you have one or fewer cards in hand, this creature gets +2/+2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: whileYouHave(bnsB10HandAtMostOne, selfCreature),
				Apply:     pumpBy(2, 2),
			},
			menace,
		},
	})
}
