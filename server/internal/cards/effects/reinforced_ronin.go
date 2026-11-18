package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reinforced Ronin — Artifact Creature — Human Samurai {R}, 2/2 (Neon
// Dynasty, #5493):
//
//	"Haste
//	 At the beginning of your end step, return this creature to its owner's
//	 hand.
//	 Channel — {1}{R}, Discard this card: Draw a card."
//
// The end-step trigger is Fleeting Effigy's: it returns the Ronin only while
// it is still the permanent the ability came from (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5e891927-b6cb-40ae-8270-0e72820fd7f5",
		Name:            "Reinforced Ronin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Reinforced Ronin — return this creature to its owner's hand",
				Do(Instruction(stdB9ReturnThisToHand))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Channel — {1}{R}, Discard this card: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{R}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Effect:  neoB07DrawsCards(1),
		}},
	})
}
