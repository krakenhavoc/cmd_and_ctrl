package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "78d74d6a-e1aa-4d70-92c3-b5338959d2fb",
		Name:            "Horned Stoneseeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Horned Stoneseeker — create a tapped Powerstone token",
				Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1})),
			WhenThisLeaves("Horned Stoneseeker — sacrifice a Powerstone", hornedStoneseekerSacrifice),
		},
	})
}

func hornedStoneseekerSacrifice(g *game.Game, item *game.StackItem) error {
	return SacrificeChoice{
		Player:     item.Controller,
		Candidates: permanentsControlledByMatching(g, item.Controller, HasSubtype("Powerstone")),
		Question:   "Horned Stoneseeker — sacrifice a Powerstone",
	}.Apply(NewContext(g, item))
}
