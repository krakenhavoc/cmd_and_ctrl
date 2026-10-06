package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curator of Destinies — Creature — Sphinx {3}{U}{U}, 4/4:
//
//	"This spell can't be countered.
//	 Flying
//	 When this creature enters, look at the top five cards of your
//	 library and separate them into a face-down pile and a face-up
//	 pile. An opponent chooses one of those piles. Put that pile into
//	 your hand and the other into your graveyard."
//
// Riddles in the Dark's split over five cards on an enters trigger
// (face_down_piles.go); the uncounterable rider is
// Spec.CantBeCountered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9938d178-0ce6-45c0-b317-fd5c54231579",
		Name:            "Curator of Destinies",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Curator of Destinies — look at the top five cards, separate them into a face-down and a face-up pile",
				func(g *game.Game, item *game.StackItem) error {
					return youSeparateFaceDownPiles(NewContext(g, item), "Curator of Destinies", 5)
				}),
		},
	})
}
