package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Second Guess — Instant {1}{U}:
//
//	"Counter target spell that's the second spell cast this turn."
//
// The clause is a target restriction, judged at announce and again at
// resolution (CR 608.2b). "Second" counts every player's spells.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b036eac6-60db-48ca-b6e1-b42afac959bd",
		Name:         "Second Guess",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell that's the second spell cast this turn", effects.SecondSpellCastThisTurn()),
		OnResolve:    effects.CounterTheTargetSpell,
	})
}
