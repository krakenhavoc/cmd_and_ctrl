package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slice from the Shadows — Instant {X}{B} (Murders at Karlov Manor,
// #4469):
//
//	"This spell can't be countered. (This includes by the ward ability.)
//	 Target creature gets -X/-X until end of turn."
//
// Spec.CantBeCountered (Abrupt Decay's), so ward's counter does nothing
// to it either, and Battle at the Bridge's -X/-X on the still-legal
// target (CR 608.2b), X the announced X (CR 107.3a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4d31d3f4-4748-4713-808f-24699fcd7b10",
		Name:            "Slice from the Shadows",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		XMatters:        true,
		Purpose:         ForTargets(HarmUntilEOT(0, 0, 0)),
		Targets:         TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			x := ctx.X()
			return vsPumpTheTarget(ctx, -x, -x, "Slice from the Shadows — -X/-X")
		},
	})
}
