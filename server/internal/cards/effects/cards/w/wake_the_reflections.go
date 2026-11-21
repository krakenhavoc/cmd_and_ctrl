package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wake the Reflections — Sorcery {W}:
//
//	"Populate. (Create a token that's a copy of a creature token you
//	 control.)"
//
// The whole card is the Populate primitive (populate.go): choose one
// of your creature tokens and copy it, with nothing to choose and
// nothing created when you control none (CR 701.36b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "211868ca-662b-4054-bf50-c9e16bb58c49",
		Name:         "Wake the Reflections",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Populate{}.Apply(ctx)
		},
	})
}
