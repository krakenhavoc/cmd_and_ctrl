package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Howl from Beyond — Instant {X}{B} (Dominaria Remastered, #5948):
//
//	"Target creature gets +X/+0 until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "403cf6ae-48a9-4ea9-894e-7135cfca4e1b",
		Name:         "Howl from Beyond",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      effects.ForTargets(effects.PowerXUntilEOT(0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.BoostUntilEOT{Target: item.Targets[0].ID, Power: ctx.X(), Label: "Howl from Beyond — +X/+0"}.Apply(ctx)
		},
	})
}
