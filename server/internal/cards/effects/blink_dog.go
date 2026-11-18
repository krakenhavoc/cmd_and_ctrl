package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blink Dog — Creature — Dog {2}{W}, 1/1:
//
//	"Double strike
//	 Teleport — {3}{W}: This creature phases out. (Treat it and anything
//	 attached to it as though they don't exist until your next turn.)"
//
// Double strike is an engine keyword carried as printed data. The ability
// is thisPhasesOut: the Dog phases out with everything attached (CR 702.26b),
// leaves combat (CR 506.4) and phases in during its controller's next untap
// step (CR 702.26a). "Teleport" is an ability word with no rules meaning
// (CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "02a7d95a-0a45-4aa3-b5b8-01c0b36462d8",
		Name:            "Blink Dog",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Activated: []ActivatedAbility{{
			Label:   "Teleport — {3}{W}: This creature phases out.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{3}{W}"),
			Effect:  thisPhasesOut,
		}},
	})
}
