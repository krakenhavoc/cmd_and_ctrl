package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simic Ragworm — Creature — Worm {3}{G}, 3/3 (Dissension, #5598):
//
//	"{U}: Untap this creature."
//
// Blizzard Elemental's untap row for {U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d8c4172-eb14-47a4-993a-6d68906a26c5",
		Name:         "Simic Ragworm",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			UntapSelf: true,
			Label:     "{U}: Untap this creature.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      ManaCost("{U}"),
			Effect:    b31UntapSelf,
		}},
	})
}
