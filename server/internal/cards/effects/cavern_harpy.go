package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cavern Harpy — Creature — Harpy Beast {U}{B}, 2/1 (Planeshift, #5615):
//
//	"Flying
//	 When this creature enters, return a blue or black creature you
//	 control to its owner's hand.
//	 Pay 1 life: Return this creature to its owner's hand."
//
// Flying is a printed keyword; the enters trigger is the shared gating
// return (gating.go), chosen as it resolves, the Harpy itself included.
// The life is the whole cost, so the self-bounce is at instant speed and
// as often as you can pay (CR 118.3). A Harpy that has left the
// battlefield by the time the ability resolves stays where it is
// (returnThisPermanentToOwnersHand).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "81c40bd8-b989-41e8-9527-eae8fb87311d",
		Name:            "Cavern Harpy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{gatingReturn("Cavern Harpy", "U", "B")},
		Activated: []ActivatedAbility{{
			Label:   "Pay 1 life: Return this creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    PayLife(1),
			Effect:  returnThisPermanentToOwnersHand,
		}},
	})
}
