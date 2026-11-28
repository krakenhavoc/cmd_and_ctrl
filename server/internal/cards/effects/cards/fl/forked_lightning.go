package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forked Lightning — Sorcery {3}{R} (Portal, S82 OLD batch 782, #5630):
//
//	"Forked Lightning deals 4 damage divided as you choose among one, two, or
//	 three target creatures."
//
// Roil's Retribution's division over one to three creatures: announced with
// the targets (CR 601.2d), each target getting at least 1; a target that is
// illegal on resolution is skipped and its share is lost (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66107cfd-4bdb-4266-a650-940743555ea4",
		Name:         "Forked Lightning",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one, two, or three target creatures").WithCount(1, 3).Dividing(effects.Divide(4)),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
