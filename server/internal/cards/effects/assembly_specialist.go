package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assembly Specialist — Artifact Creature — Borg {1}{W}, 1/1
// (Star Trek, #4973):
//
//	"When this creature enters, create a 1/1 colorless Borg artifact
//	 creature token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "deb26a1b-2b83-4da2-a2e5-31c17a57eb1e",
		Name:         "Assembly Specialist",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Assembly Specialist — create a 1/1 colorless Borg artifact creature token",
				Do(CreateToken{Template: TokenCard("1/1 colorless Borg artifact"), N: 1})),
		},
	})
}
