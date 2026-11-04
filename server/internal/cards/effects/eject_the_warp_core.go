package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eject the Warp Core — Instant {1}{B}:
//
//	"Kicker—Sacrifice an artifact or creature. (You may sacrifice an
//	 artifact or creature in addition to any other costs as you cast this
//	 spell.)
//	 Target creature gets -2/-2 until end of turn. If this spell was
//	 kicked, that creature gets -6/-6 until end of turn instead."
//
// Falling Timber's sacrifice kicker (CR 702.33a): the sacrifice is an
// additional cost paid as the spell is cast. Kicked, the one target gets
// -6/-6 in place of the -2/-2, never both.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "c06c17c2-203c-4b4a-8564-f29bc41c9eb5",
		Name:          "Eject the Warp Core",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{KickerSacrifice("an artifact or creature", Or(Artifact(), Creature()))},
		Targets:       TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2
			if ctx.WasKicked() {
				n = 6
			}
			for _, t := range ctx.LegalTargets() {
				return BoostUntilEOT{Target: t.ID, Power: -n, Toughness: -n, Label: "Eject the Warp Core"}.Apply(ctx)
			}
			return nil
		},
	})
}
