package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mutant's Prey — Instant {G} (Dragon's Maze, S82 OLD b405, #5564):
//
//	"Target creature you control with a +1/+1 counter on it fights target
//	 creature an opponent controls."
//
// Prey Upon's fight on two clauses, the first narrowed to a creature that
// has a +1/+1 counter on it. Each clause is re-checked at resolution, so a
// creature that lost its counter in response is an illegal target (CR 608.2b)
// and no damage is dealt (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3836cf33-bd1d-485d-94a2-85ea210ddf7a",
		Name:         "Mutant's Prey",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control with a +1/+1 counter on it", effects.YouControl(),
				func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
					return effects.HasCounterOfKind(c, game.CounterPlusOne)
				}),
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		),
		OnResolve: effects.ResolveSteps(effects.Instruction(effects.FightTargetClauses)),
	})
}
