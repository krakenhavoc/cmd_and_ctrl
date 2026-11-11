package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sphinx of Enlightenment — Creature — Sphinx {4}{U}{U}, 5/5 (Bloomburrow
// Commander, #5368):
//
//	"Flying
//	 When this creature enters, target opponent draws a card and you draw
//	 three cards."
//
// The opponent is chosen as the trigger goes on the stack and draws
// first, as printed. If they are no longer a legal target on resolution
// they draw nothing and the controller still draws three.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e6284fb3-7cf8-4730-b156-10085b70b0e8",
		Name:            "Sphinx of Enlightenment",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Sphinx of Enlightenment — target opponent draws a card and you draw three cards",
				cmdr2S22TargetPlayerDrawsThenYouDraw(1, 3)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}
