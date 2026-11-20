package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Might of Murasa — Instant {1}{G} (Zendikar Rising, #5508):
//
//	"Kicker {2}{G} (You may pay an additional {2}{G} as you cast this
//	 spell.)
//	 Target creature gets +3/+3 until end of turn. If this spell was
//	 kicked, that creature gets +5/+5 until end of turn instead."
//
// Colossal Growth's shape: one pump, its size read off the cast
// (ctx.WasKicked). The "instead" replaces the +3/+3, so a kicked cast is
// +5/+5 and never +8/+8.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "902eb71f-62b3-40cb-816b-f6b7e92e91e7",
		Name:          "Might of Murasa",
		Completeness:  CompletenessFull,
		Purpose:       ForTargets(BenefitUntilEOT(0, 3, 3)),
		Targets:       TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{Kicker("{2}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			n := 3
			if ctx.WasKicked() {
				n = 5
			}
			return BoostUntilEOT{Target: id, Power: n, Toughness: n, Label: "Might of Murasa — pump"}.Apply(ctx)
		},
	})
}
