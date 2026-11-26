package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Panic Attack — Sorcery {2}{R} (Magic 2010, #5585):
//
//	"Up to three target creatures can't block this turn."
//
// Abandon the Post's body with three targets: each creature still legal as
// it resolves can't block until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1346677-d89e-49d0-9819-08805f00bb64",
		Name:         "Panic Attack",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.OldB499CantBlockTargets(ctx, "Panic Attack")
		},
	})
}
