package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vampire Spawn — Creature — Vampire {2}{B}, 2/3 (Foundations, #4465):
//
//	"When this creature enters, each opponent loses 2 life and you gain
//	 2 life."
//
// Skirmish Rhino's drain on a black body: every opponent loses 2 life,
// then you gain 2, a fixed amount however many opponents there are.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33cee4b4-3d80-46f1-9223-49ed785891a1",
		Name:         "Vampire Spawn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Vampire Spawn — each opponent loses 2 life and you gain 2 life",
				Do(EachOpponentLosesLife{N: 2}, GainLife{Amount: 2})),
		},
	})
}
