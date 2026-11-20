package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marrow Bats — {4}{B} Creature — Bat Skeleton 4/1 (Avacyn Restored, #5570):
//
//	"Flying
//	 Pay 4 life: Regenerate this creature."
//
// Flying and a regeneration row paid in life, Mischievous
// Poltergeist's shape at 4 life (CR 701.19a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f20d96cc-3949-4888-a909-c7d67c4d31c7",
		Name:            "Marrow Bats",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Pay 4 life: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    PayLife(4),
			Effect:  regenerateThisSource,
		}},
	})
}
