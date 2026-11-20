package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Homarid Explorer — Creature — Homarid Scout {3}{U}, 3/3 (Dominaria, #5530):
//
//	"When this creature enters, target player mills four cards."
//
// Wall of Lost Thoughts' enters trigger: target player mills four.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ac30745c-addf-433a-be73-d3d9f4cc4328",
		Name:         "Homarid Explorer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Homarid Explorer — target player mills four cards", TargetPlayerMills(4)),
				TargetPlayer("target player")),
		},
	})
}
