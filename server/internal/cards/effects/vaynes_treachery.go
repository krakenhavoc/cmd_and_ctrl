package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vayne's Treachery — Instant {1}{B} (Final Fantasy, #4462):
//
//	"Kicker—Sacrifice an artifact or creature. (You may sacrifice an
//	 artifact or creature in addition to any other costs as you cast
//	 this spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was
//	 kicked, that creature gets -6/-6 until end of turn instead."
//
// Stomped by the Foot's shape with a bigger kicked shrink: a non-mana
// kicker paid as the spell is cast (CR 702.33), and "instead" is one
// read of the kick at resolution, so the target gets one shrink of
// whichever size, never both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "93af8f9d-6877-4cb8-a744-c89186db22e1",
		Name:          "Vayne's Treachery",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{KickerSacrifice("an artifact or creature", Or(Artifact(), Creature()))},
		Targets:       TargetCreature("target creature"),
		Purpose:       ForTargets(HarmUntilEOT(0, -2, -2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.WasKicked() {
				return vsPumpTheTarget(ctx, -6, -6, "Vayne's Treachery — kicked: -6/-6")
			}
			return vsPumpTheTarget(ctx, -2, -2, "Vayne's Treachery — -2/-2")
		},
	})
}
