package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Timestream Navigator — Creature — Human Pirate Wizard {1}{U}, 1/1:
//
//	"Ascend (If you control ten or more permanents, you get the city's
//	 blessing for the rest of the game.)
//	 {2}{U}{U}, {T}, Put this creature on the bottom of its owner's
//	 library: Take an extra turn after this one. Activate only if you
//	 have the city's blessing."
//
// The put-on-the-bottom is a cost (CR 602.2b), paid at announce through
// game.AbilityCost.BottomSelf (#2726): the creature is in the library
// before anyone can respond, so it is gone even if the ability is
// countered. The extra turn is Capture of Jingzhou's (TakeExtraTurn);
// it does not read the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d0896996-ee31-402e-b642-4e7cca4929fe",
		Name:            "Timestream Navigator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordAscend},
		Activated: []ActivatedAbility{{
			Label:     "{2}{U}{U}, {T}, Put this creature on the bottom of its owner's library: Take an extra turn after this one. Activate only if you have the city's blessing.",
			Cost:      Plus(ManaCost("{2}{U}{U}"), TapCost(), PutThisOnTheBottomOfItsOwnersLibrary()),
			Condition: YouHaveTheCitysBlessingCondition(),
			// ADR 0142: an extra turn answers nothing on the stack.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect:  youTakeAnExtraTurnEffect,
		}},
	})
}
