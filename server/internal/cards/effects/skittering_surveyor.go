package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skittering Surveyor — Artifact Creature — Construct {3}, 1/2:
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// The "you may" is the search prompt's decline, which skips the shuffle
// too; the pick is revealed and goes to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "85364398-923c-4ee9-8519-576fce1c26f7",
		Name:         "Skittering Surveyor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Skittering Surveyor — search for a basic land card", mayFetchABasicLandToHand("Skittering Surveyor")),
		},
	})
}
