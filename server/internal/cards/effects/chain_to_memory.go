package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chain to Memory — Instant {U} (Theros Beyond Death, #5514):
//
//	"Target creature gets -4/-0 until end of turn. Scry 2."
//
// Depower's -4/-0 on the target, then the scry.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3805bb3-544c-4142-a6df-c60fa38d625a",
		Name:         "Chain to Memory",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose: game.Purpose{
			Targets: game.ForTargets(HarmUntilEOT(0, -4, 0)),
			Scry:    2,
		},
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return BoostUntilEOT{Target: id, Power: -4, Label: "Chain to Memory — -4/-0 until end of turn"}.Apply(ctx)
			}),
			Scry{N: 2},
		),
	})
}
