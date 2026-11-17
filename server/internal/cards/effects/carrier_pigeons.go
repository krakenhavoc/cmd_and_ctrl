package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carrier Pigeons — Creature — Bird {3}{W}, 1/1 (Alliances, #5634):
//
//	"Flying
//	 When this creature enters, draw a card at the beginning of the next
//	 turn's upkeep."
//
// Flying is an engine keyword carried as printed data (CR 702.9b). The
// enters trigger uses the stack (CR 603.6a); when it resolves it
// schedules the delayed draw at the next upkeep at the table, whoever's
// turn it is (CR 603.7a, next_upkeep_draw.go). The draw happens even if
// the Pigeons have left by then.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2d30a634-43c6-4e78-ba10-7ef7ee17ac40",
		Name:            "Carrier Pigeons",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Carrier Pigeons — draw a card at the beginning of the next turn's upkeep",
				Do(drawACardAtTheNextUpkeep("Carrier Pigeons — draw a card"))),
		},
	})
}
