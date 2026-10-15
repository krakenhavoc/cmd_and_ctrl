package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wormfang Manta — Creature — Nightmare Fish Beast {5}{U}{U}, 6/1:
//
//	"Flying
//	 When this creature enters, you skip your next turn.
//	 When this creature leaves the battlefield, you take an extra turn
//	 after this one."
//
// Two triggers on the stack. The skip is a debt spent when the
// controller's next turn would begin, normal or extra (CR 614.10, ADR
// 0188 Decision 1, #3565), so a Manta that leaves the battlefield during
// the controller's own turn queues the extra turn, and the skip is then
// spent on that extra turn: the pair cancels, as the Magosi ruling says.
// Left on the battlefield, the controller loses the turn and the Manta
// is simply a 6/1 flier for seven. Leaving is any exit, not only dying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "17abf8c6-70b6-4877-a0f2-a87b22fb2828",
		Name:            "Wormfang Manta",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{SkipsTurns: 1, ExtraTurns: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wormfang Manta — you skip your next turn", Do(SkipYourNextTurns{})),
			WhenThisLeaves("Wormfang Manta — you take an extra turn after this one", Do(TakeExtraTurn{})),
		},
	})
}
