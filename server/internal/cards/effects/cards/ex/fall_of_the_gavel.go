package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fall of the Gavel — Instant {3}{W}{U} (Return to Ravnica, #5566):
//
//	"Counter target spell. You gain 5 life."
//
// One target: when the spell is gone on resolution this does nothing at all
// (CR 608.2b), so no life is gained.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ad33099-0392-4f68-a4c5-7fa3ef2d473e",
		Name:         "Fall of the Gavel",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, LifeGain: 5},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B35CounterChosenSpellThenGainLife(item, ctx, 5)
		},
	})
}
