package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Farsight Adept — Creature — Kor Wizard {2}{W}, 3/3 (Zendikar Rising,
// #5508):
//
//	"When this creature enters, you and target opponent each draw a
//	 card."
//
// Flumph's draw on an enters trigger: the opponent is chosen as the
// trigger goes on the stack; you draw, then the opponent does if still a
// legal target (youAndTargetOpponentEachDraw).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72b10cba-0f46-4b99-9634-e7173ac4c259",
		Name:         "Farsight Adept",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Farsight Adept — you and target opponent each draw a card", youAndTargetOpponentEachDraw()),
				TargetPlayer("target opponent", Opponent()),
			),
		},
	})
}
