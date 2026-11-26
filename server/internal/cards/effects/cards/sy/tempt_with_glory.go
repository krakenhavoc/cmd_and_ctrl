package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tempt with Glory — Sorcery {5}{W} (Commander 2013, S82 older sets):
//
//	"Tempting offer — Put a +1/+1 counter on each creature you control. Each
//	 opponent may put a +1/+1 counter on each creature they control. For each
//	 opponent who does, put a +1/+1 counter on each creature you control."
//
// The shared tempting offer (NewCmdr2S49Tempting): the caster puts a counter
// on each creature they control, each opponent in turn order is asked, and an
// opponent who accepts puts one on each of their own and the caster does it
// again. Each set of creatures is read as that player's counters are put on.
//
// No simplification.
func init() {
	tempt := effects.NewCmdr2S49Tempting("confirm/tempt-with-glory",
		"Tempt with Glory — put a +1/+1 counter on each creature you control (the caster does it again if you do)?",
		func(ctx *effects.Context, player uuid.UUID) error {
			return effects.OldB381PlusOneOnEachCreature(ctx, player)
		})
	effects.Register(effects.Spec{
		OracleID:     "5a8dd1b7-b63e-4997-9fe8-5e8816bc051b",
		Name:         "Tempt with Glory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return tempt.Resolve(ctx)
		},
	})
}
