package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rootwater Alligator — Creature — Crocodile {3}{G}, 3/2
// (Exodus, #5626):
//
//	"Sacrifice a Forest: Regenerate this creature."
//
// The shared regenerateThisSource body (CR 701.19a) behind a one-land
// sacrifice cost, Goblin Chirurgeon's Subtype clause: any permanent you
// control with the Forest land type will do, a nonbasic Forest
// included, and nothing else. The sacrifice is paid as it is activated
// (CR 602.2b), so it needs a Forest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2bf84938-6049-4d03-b09e-85bd0763a187",
		Name:         "Rootwater Alligator",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice a Forest: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    SacrificeN(1, "a Forest", Subtype("Forest")),
			Effect:  regenerateThisSource,
		}},
	})
}
