package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dream Beavers — {B} Creature — Beaver Nightmare 1/1 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"Flying
//	 When this creature enters, each opponent loses 1 life and you gain
//	 1 life. Scry 1."
//
// One enters trigger, its instructions in printed order (CR 608.2c): the
// drain, then the scry, which is last so nothing waits on its prompt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a1cec8e8-cf3c-42a4-b625-ca0cd148664e",
		Name:            "Dream Beavers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dream Beavers — each opponent loses 1 life, you gain 1 life, scry 1",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}, Scry{N: 1})),
		},
	})
}
