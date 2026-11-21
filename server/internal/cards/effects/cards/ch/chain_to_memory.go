package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chain to Memory — Instant {U} (Theros Beyond Death, #5514):
//
//	"Target creature gets -4/-0 until end of turn. Scry 2."
//
// Depower's -4/-0 on the target, then the scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3805bb3-544c-4142-a6df-c60fa38d625a",
		Name:         "Chain to Memory",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.HarmUntilEOT(0, -4, 0)),
			Scry:    2,
		},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: -4, Label: "Chain to Memory — -4/-0 until end of turn"}.Apply(ctx)
			}),
			effects.Scry{N: 2},
		),
	})
}
