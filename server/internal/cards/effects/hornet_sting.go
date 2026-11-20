package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hornet Sting — Instant {G} (Magic 2011, #5579):
//
//	"Hornet Sting deals 1 damage to any target."
//
// Searing Touch's 1 damage without the buyback: a creature, a
// planeswalker, a battle or a player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "25edeff3-7727-4b80-9459-2bb6b8587fb9",
		Name:         "Hornet Sting",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 1)), Answers: game.AnswerRemove},
		OnResolve:    damageToFirstTarget(1),
	})
}
