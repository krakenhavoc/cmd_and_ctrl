package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Isperia's Skywatch — Creature — Vedalken Knight {5}{U}, 3/3 (Return to
// Ravnica, #5566):
//
//	"Flying
//	 When this creature enters, detain target creature an opponent
//	 controls. (Until your next turn, that creature can't attack or block
//	 and its activated abilities can't be activated.)"
//
// Azorius Arrester's trigger on a flyer: the target is chosen as the
// trigger goes on the stack and detained as it resolves (CR 701.35a),
// until its controller's next turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5101fa05-3cca-499c-8c48-198ba0f8c672",
		Name:            "Isperia's Skywatch",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Isperia's Skywatch — detain target creature an opponent controls",
				detainEachLegalTarget("Isperia's Skywatch — detain")),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
