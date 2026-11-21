package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Death Wind — Instant {X}{B} (Jumpstart 2022, #5948):
//
//	"Target creature gets -X/-X until end of turn."
//
// Black Sun's Twilight's -X/-X: X is announced with the cast, and a target
// that is illegal as the spell resolves leaves it with no effect (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a636b62a-340b-444d-9161-f6d3367b745c",
		Name:         "Death Wind",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GapB170MinusXMinusX(ctx, ctx.X())
		},
	})
}
