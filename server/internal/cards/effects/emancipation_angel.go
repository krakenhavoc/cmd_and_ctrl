package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emancipation Angel — {1}{W}{W} Creature — Angel 3/3 (Avacyn Restored, #5570):
//
//	"Flying
//	 When this creature enters, return a permanent you control to its
//	 owner's hand."
//
// Flying and Whitemane Lion's enters return, for any permanent: the
// choice is made as the trigger resolves, not targeted, and the Angel
// itself may be the one returned.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3834d3b0-9b66-465a-a8dc-22875a819fd9",
		Name:            "Emancipation Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Emancipation Angel — return a permanent you control", Do(ReturnOneYouControl{
				Question: "Emancipation Angel — return a permanent you control to its owner's hand",
			})),
		},
	})
}
