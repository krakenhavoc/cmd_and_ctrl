package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oltec Matterweaver — Creature — Human Artificer {2}{W}, 3/2 (The Big
// Score, #4468):
//
//	"Whenever you cast a creature spell, choose one —
//	 • Create a 1/1 colorless Gnome artifact creature token.
//	 • Create a token that's a copy of target artifact token you control."
//
// A modal cast trigger; the copy mode is only choosable with an artifact
// token to target (CR 700.2b) and copies what the token was created as
// (CR 707.2).
//
// No simplification.
func init() {
	t := WheneverYouCast(Creature(), "Oltec Matterweaver — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeSteps("Create a 1/1 colorless Gnome artifact creature token.", nil,
			CreateToken{Template: TokenCard("1/1 colorless Gnome artifact"), N: 1}),
		ModeDoing("Create a token that's a copy of target artifact token you control.", b35ArtifactTokenYouControl(), b35CopyTargetTokenModeBody),
	)
	Register(Spec{
		OracleID:     "3cb91696-242f-4e91-806b-61b1bb45a20c",
		Name:         "Oltec Matterweaver",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
