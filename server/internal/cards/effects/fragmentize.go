package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fragmentize — Sorcery {W} (Kaladesh, #5539):
//
//	"Destroy target artifact or enchantment with mana value 4 or less."
//
// Naturalize with a mana value bound on the target, judged at announce
// and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f025c1da-734e-4cbb-99d8-9d700d6edbde",
		Name:         "Fragmentize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target artifact or enchantment with mana value 4 or less", Or(Artifact(), Enchantment()), ManaValueLE(4)),
		OnResolve:    destroyTheTargetPermanent,
	})
}
