package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Absorb — Instant {W}{U}{U} (EDHREC rank 3716):
//
//	"Counter target spell. You gain 3 life."
//
// Cancel with a rider, the Azorius mirror of Countersquall. The
// announced spell is countered if it is still on the stack, and the
// caster gains 3 either way — the life is not conditional on the
// counter taking, so a spell that can't be countered still pays it
// out, as printed. With the target gone before resolution the whole
// spell fizzles (CR 608.2b) and no life is gained, which is also as
// printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "132ca99a-a3c7-4ed6-b4d0-0edcd7140ca2",
		Name:         "Absorb",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: a counterspell (the enumerator reads the counter itself; remove, as sweep ruling 8's stack removal)
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B35CounterChosenSpellThenGainLife(item, ctx, 3)
		},
	})
}
