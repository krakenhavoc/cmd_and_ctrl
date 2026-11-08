package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glidedive Duo — Creature — Bat Lizard {4}{B}, 3/3 (Bloomburrow,
// #4467):
//
//	"Flying
//	 When this creature enters, each opponent loses 2 life and you gain 2
//	 life."
//
// Skirmish Rhino's drain on a flier: every opponent loses 2 life, then
// you gain 2, a fixed amount however many opponents there are.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33c06c11-8d0a-45ce-adea-c9a7a8ac8779",
		Name:            "Glidedive Duo",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Glidedive Duo — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
