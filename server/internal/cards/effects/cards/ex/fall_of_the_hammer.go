package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fall of the Hammer — Instant {1}{R} (Commander 2015, S82 OLD #5546):
//
//	"Target creature you control deals damage equal to its power to another
//	 target creature."
//
// Bite Down's one-sided bite over two clauses, the second Distinct from the
// first ("another"): slot 0 deals, slot 1 takes. The damage is the biter's
// power as it resolves, dealt by the biter itself, and nothing is dealt unless
// both picks are still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e3fe8e4-e56a-4109-a314-96c44bbe7d22",
		Name:         "Fall of the Hammer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("another target creature")),
		),
		OnResolve: effects.OneSidedBite,
	})
}
