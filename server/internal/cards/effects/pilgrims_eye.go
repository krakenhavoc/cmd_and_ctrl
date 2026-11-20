package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pilgrim's Eye — Artifact Creature — Thopter {3}, 1/1 (Worldwake, #5582):
//
//	"Flying
//	 When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// Sylvan Ranger's enters trigger on a flyer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "66155010-cc2c-449a-85d5-92c95aade514",
		Name:            "Pilgrim's Eye",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tutors: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pilgrim's Eye — search for a basic land card", mayFetchABasicLandToHand("Pilgrim's Eye")),
		},
	})
}
