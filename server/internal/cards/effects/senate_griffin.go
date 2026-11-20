package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Senate Griffin — Creature — Griffin {2}{W/U}{W/U}, 3/2 (Ravnica
// Allegiance, #5521):
//
//	"Flying
//	 When this creature enters, scry 1."
//
// Silver Raven's two lines: flying is printed data (CR 702.9b), and the
// enters trigger scries 1 (CR 701.22a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7ddf9c12-8f75-481c-a218-d42f925c6503",
		Name:            "Senate Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Senate Griffin — scry 1", Do(Scry{N: 1})),
		},
	})
}
