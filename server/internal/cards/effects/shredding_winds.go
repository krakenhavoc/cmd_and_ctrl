package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shredding Winds — Instant {2}{G} (Theros, #5561):
//
//	"Shredding Winds deals 7 damage to target creature with flying."
//
// Seven damage to a creature with flying, read as it is cast and again
// as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f914186-082e-4a85-a1cc-e9788074637b",
		Name:         "Shredding Winds",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 7)), Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature with flying", HasKeyword("flying")),
		OnResolve:    damageToFirstTarget(7),
	})
}
