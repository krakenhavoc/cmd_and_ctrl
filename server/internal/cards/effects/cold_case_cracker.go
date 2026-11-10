package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cold Case Cracker — Creature — Spirit Detective {3}{U}, 3/3 (Murders
// at Karlov Manor, #4469):
//
//	"Flying
//	 When this creature dies, investigate. (Create a Clue token. It's an
//	 artifact with "{2}, Sacrifice this token: Draw a card.")"
//
// Flying rides PrintedKeywords; the dies trigger makes one Clue for the
// Cracker's controller as it last existed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2e18ee5b-cda3-428b-be91-939c410256ac",
		Name:            "Cold Case Cracker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Cold Case Cracker — investigate", Do(Investigate{N: 1})),
		},
	})
}
