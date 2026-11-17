package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Final Flourish — Instant {1}{B} (March of the Machine, S82 #5477):
//
//	"Kicker—Sacrifice an artifact or creature. (You may sacrifice an
//	 artifact or creature in addition to any other costs as you cast this
//	 spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was kicked,
//	 that creature gets -6/-6 until end of turn instead."
//
// The sacrifice is the kicker (CR 702.33), paid on casting; kicked, the
// creature gets -6/-6 INSTEAD of -2/-2, never both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "63a85361-56bc-4209-aea8-b49319bc7b92",
		Name:          "Final Flourish",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{KickerSacrifice("an artifact or creature", Or(Artifact(), Creature()))},
		Targets:       TargetCreature("target creature"),
		Purpose:       ForTargets(HarmUntilEOT(0, -2, -2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.WasKicked() {
				return vsPumpTheTarget(ctx, -6, -6, "Final Flourish — kicked: -6/-6")
			}
			return vsPumpTheTarget(ctx, -2, -2, "Final Flourish — -2/-2")
		},
	})
}
