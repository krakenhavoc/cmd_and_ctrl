package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bombard — Instant {2}{R} (Edge of Eternities, #4461):
//
//	"Bombard deals 4 damage to target creature."
//
// Flame Slash's body at instant speed: 4 damage to one creature,
// whoever controls it, skipped if the target is no longer legal as
// Bombard resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0899c12-5462-4ff1-a9e5-d99cffa7c6cc",
		Name:         "Bombard",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(DamageToTarget(0, 4)),
		},
		OnResolve: damageToFirstTarget(4),
	})
}
