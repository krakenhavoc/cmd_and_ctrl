package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drownyard Explorers — Creature — Human Wizard {3}{U}, 2/4
// (Shadows over Innistrad, #5543):
//
//	"When this creature enters, investigate. (Create a Clue token. It's
//	 an artifact with "{2}, Sacrifice this token: Draw a card.")"
//
// Novice Inspector's enters trigger: investigate (CR 701.16).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c1ebbe89-8976-4ed7-a68a-8b08aac4616c",
		Name:         "Drownyard Explorers",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Drownyard Explorers — investigate", Do(Investigate{N: 1})),
		},
	})
}
