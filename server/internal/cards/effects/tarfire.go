package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tarfire — Kindred Instant — Goblin {R} (Lorwyn, #5593):
//
//	"Tarfire deals 2 damage to any target."
//
// Shock as a Kindred Goblin card (CR 308.1): cast and resolved as an
// instant, and a Goblin card for anything that counts them in a hand or a
// graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f07b821c-bbc1-4def-b009-6c03689478d0",
		Name:         "Tarfire",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(2),
	})
}
