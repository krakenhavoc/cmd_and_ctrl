package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "2a8e14be-6f93-428d-a204-e2399dc93797",
		Name:            "Shrieking Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Shrieking Drake — return a creature you control", effects.Do(effects.ReturnOneYouControl{
				Match:    effects.MatchCreature,
				Question: "Shrieking Drake — return a creature you control to its owner's hand",
			})),
		},
	})
}
