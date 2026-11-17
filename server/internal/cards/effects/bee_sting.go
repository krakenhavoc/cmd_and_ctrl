package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bee Sting — Sorcery {3}{G} (Portal, #5630):
//
//	"Bee Sting deals 2 damage to any target."
//
// Shock's body as a sorcery (damageToFirstTarget). A target that
// became illegal is never dealt damage (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f637d525-2f29-488a-9269-8e5aa377fbb7",
		Name:         "Bee Sting",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 2)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(2),
	})
}
