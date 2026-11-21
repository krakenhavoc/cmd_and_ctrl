package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Valor Made Real — Instant {W}:
//
//	"Target creature can block any number of creatures this turn."
//
// BlockCapacityUntilEOT with AnyNumber (#1715).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bd1348c-51cc-4426-b995-8bb6d4a5bd2e",
		Name:         "Valor Made Real",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			return effects.BlockCapacityUntilEOT{
				Target:    ts[0].ID,
				AnyNumber: true,
				Label:     "Valor Made Real — can block any number of creatures",
			}.Apply(ctx)
		},
	})
}
