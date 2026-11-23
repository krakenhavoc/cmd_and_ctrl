package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Untamed Might — Instant {X}{G} (Scars of Mirrodin, #5577):
//
//	"Target creature gets +X/+X until end of turn."
//
// Enrage's shape on both sides: X is the announced X (CR 107.3a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b6dd710-f7f2-4fc8-9b66-b48c7737b135",
		Name:         "Untamed Might",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      effects.ForTargets(effects.PumpXUntilEOT(0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BoostUntilEOT{Target: id, Power: ctx.X(), Toughness: ctx.X(), Label: "Untamed Might — +X/+X until end of turn"}.Apply(ctx)
		},
	})
}
