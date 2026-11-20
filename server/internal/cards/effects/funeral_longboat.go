package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Funeral Longboat — Artifact — Vehicle {2}, 3/3 (Kaldheim, #5505):
//
//	"Vigilance
//	 Crew 1"
//
// Brute Suit's text exactly: vigilance rides PrintedKeywords and crew is
// the shared CrewCost / CrewEffect row (CR 702.122).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a994de7b-0128-4f03-922f-c8cd80cb5852",
		Name:            "Funeral Longboat",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Funeral Longboat"),
		}},
	})
}
