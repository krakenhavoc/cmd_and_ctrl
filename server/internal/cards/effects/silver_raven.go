package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Silver Raven — Artifact Creature — Bird {U}, 1/1 (Adventures in the Forgotten Realms):
//
//	Flying
//	 When this creature enters, scry 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4da64806-8a1d-4926-a168-695df43309cf",
		Name:            "Silver Raven",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Silver Raven — scry 1", Do(Scry{N: 1})),
		},
	})
}
