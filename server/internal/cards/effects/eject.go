package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eject — Instant {3}{U} (Final Fantasy, #4462):
//
//	"This spell can't be countered.
//	 Return target nonland permanent to its owner's hand.
//	 Draw a card."
//
// Leave in the Dust's body (bounceTargetThenDrawACard) with
// Spec.CantBeCountered, which functions on the stack (CR 113.6g). A
// target that became illegal in response makes the whole spell do
// nothing, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9a18fb98-796a-43b1-8157-e3afca53a1c0",
		Name:            "Eject",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		Purpose:         game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:         TargetPermanent("target nonland permanent", Nonland()),
		OnResolve:       bounceTargetThenDrawACard,
	})
}
