package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undo — Sorcery {1}{U}{U} (Visions, #5631):
//
//	"Return two target creatures to their owners' hands."
//
// Exactly two different creatures (CR 115.3). Each one still legal as it
// resolves goes to its owner's hand, and one that left in response is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b446c3bd-3c7d-4d59-ba59-3dd80818465e",
		Name:         "Undo",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("two target creatures").WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}
