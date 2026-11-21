package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horned Stoneseeker — Creature — Lizard {1}{R}, 2/2 (The Brothers'
// War, #5482):
//
//	"Menace
//	 When this creature enters, create a tapped Powerstone token.
//	 When this creature leaves the battlefield, sacrifice a Powerstone."
//
// Two triggers, one per printed ability. The leave trigger is
// mandatory and the controller chooses which Powerstone goes (any
// Powerstone they control, token or not); with none it does nothing
// (CR 701.21a). The candidates are read when the trigger resolves, so
// the Powerstone made by an enters trigger that is still waiting
// beneath it counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "78d74d6a-e1aa-4d70-92c3-b5338959d2fb",
		Name:            "Horned Stoneseeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Horned Stoneseeker — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
			effects.WhenThisLeaves("Horned Stoneseeker — sacrifice a Powerstone", hornedStoneseekerSacrifice),
		},
	})
}

func hornedStoneseekerSacrifice(g *game.Game, item *game.StackItem) error {
	return effects.SacrificeChoice{
		Player:     item.Controller,
		Candidates: effects.PermanentsControlledByMatching(g, item.Controller, effects.HasSubtype("Powerstone")),
		Question:   "Horned Stoneseeker — sacrifice a Powerstone",
	}.Apply(effects.NewContext(g, item))
}
