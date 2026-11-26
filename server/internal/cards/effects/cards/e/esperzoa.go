package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Esperzoa — Artifact Creature — Jellyfish {2}{U}, 4/3 (Duel Decks: Elspeth
// vs. Tezzeret, #5578):
//
//	"Flying
//	 At the beginning of your upkeep, return an artifact you control to its
//	 owner's hand."
//
// The upkeep trigger is mandatory: its controller picks any artifact they
// control as it resolves, and the Esperzoa, an artifact itself, is a legal
// pick (Salvage Scuttler's ReturnOneYouControl with the artifact noun).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c9a188f9-2e65-4a66-826f-0150f25730af",
		Name:            "Esperzoa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Esperzoa — return an artifact you control to its owner's hand",
				effects.Do(effects.ReturnOneYouControl{
					Match:    effects.MatchArtifact,
					Question: "Esperzoa — return an artifact you control to its owner's hand",
				})),
		},
	})
}
