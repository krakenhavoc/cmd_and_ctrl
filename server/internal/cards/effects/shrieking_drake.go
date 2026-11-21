package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shrieking Drake — Creature — Drake {U}, 1/1 (Modern Horizons 3):
//
//	"Flying
//	 When this creature enters, return a creature you control to its
//	 owner's hand."
//
// Whitemane Lion's trigger word for word: a choice made as it resolves
// (ReturnOneYouControl), not a target, and the Drake itself is always a
// candidate while it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2a8e14be-6f93-428d-a204-e2399dc93797",
		Name:            "Shrieking Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Shrieking Drake — return a creature you control", Do(ReturnOneYouControl{
				Match:    MatchCreature,
				Question: "Shrieking Drake — return a creature you control to its owner's hand",
			})),
		},
	})
}
