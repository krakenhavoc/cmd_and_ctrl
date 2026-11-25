package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Salvage Scuttler — Creature — Crab {4}{U}, 4/4 (Aether Revolt, S82 older
// sets, #5536):
//
//	"Whenever this creature attacks, return an artifact you control to its
//	 owner's hand."
//
// Not a target: the controller picks the artifact as the trigger resolves
// (CR 608.2d), and with none to return nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "01a6f8e6-439b-4de4-9f0a-21a035fd849e",
		Name:         "Salvage Scuttler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Salvage Scuttler — return an artifact you control to its owner's hand",
				effects.Do(effects.ReturnOneYouControl{
					Match:    effects.MatchArtifact,
					Question: "Salvage Scuttler — return an artifact you control to its owner's hand",
				})),
		},
	})
}
