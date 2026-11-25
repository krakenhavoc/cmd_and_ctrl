package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Peel from Reality — Instant {1}{U} (Commander Legends):
//
//	"Return target creature you control and target creature you don't control
//	 to their owners' hands."
//
// Two target clauses; each target still legal as the spell resolves goes to
// its owner's hand (CR 608.2b), so one going away does not stop the other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65dc05d0-f885-4fe5-9a23-d6b75185c179",
		Name:         "Peel from Reality",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}
