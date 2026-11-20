package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wing Puncture — Instant {G} (Scars of Mirrodin, #5577):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature with flying."
//
// Rabid Bite's one-sided fight (oneSidedBite) with a flying second
// clause. Nothing is dealt unless both targets are still legal
// (CR 608.2b); the two may be the same creature (CR 115.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5de32ec7-3085-4b6b-a52e-5a1b3fdf07d3",
		Name:         "Wing Puncture",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature with flying", HasKeyword("flying")),
		),
		OnResolve: oneSidedBite,
	})
}
