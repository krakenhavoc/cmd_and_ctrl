package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Augur il-Vec — Creature — Human Cleric {1}{W}, 1/3 (Future Sight, #5594):
//
//	"Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)
//	 Sacrifice this creature: You gain 4 life. Activate only during your
//	 upkeep."
//
// The Future Sight augur cycle: a sacrifice-itself ability with Hell's
// Caretaker's upkeep window (DuringYourUpkeep is the step AND your turn).
// The sacrifice is the cost, paid at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "34e05dc1-35af-402d-96ae-e249b3738493",
		Name:            "Augur il-Vec",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shadow"},
		Activated: []ActivatedAbility{{
			Label:     "Sacrifice this creature: You gain 4 life. Activate only during your upkeep.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      SacrificeThis(),
			Condition: DuringYourUpkeep(),
			Effect:    Do(GainLife{Amount: 4}),
		}},
	})
}
