package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Etched Familiar — Artifact Creature — Phyrexian Fox 3/2 {2}{B} (March of the Machine, S82):
//
//	"When this creature dies, each opponent loses 2 life and you gain 2 life."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e2bc064e-4739-4eb1-ba44-1fd92b2829d7",
		Name:         "Etched Familiar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Etched Familiar — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
