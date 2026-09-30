package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Widow's Bite — {1}{B} Instant:
//
//	"Teamwork 3 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 3 or more.)
//	 Choose one. If this spell was cast using teamwork, choose both
//	 instead.
//	 • Target creature gains deathtouch until end of turn.
//	 • Target creature gets -2/-2 until end of turn."
//
// #1703: Teamwork(3) with InsteadIf(2, TeamworkUsed). The two bullets
// may target the same creature or different ones (CR 601.2c), and run
// in printed order. No simplification.
func init() {
	Register(Spec{
		OracleID:      "39f2a632-30d2-4b35-9e7f-fef27713a4f7",
		Name:          "Widow's Bite",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(3)},
		Modes: ChooseOne(
			Mode("Target creature gains deathtouch until end of turn.",
				TargetCreature("target creature")),
			Mode("Target creature gets -2/-2 until end of turn.",
				TargetCreature("target creature")),
		).InsteadIf(2, TeamworkUsed),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return BulletsInPrintedOrder(item, ctx,
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}, Label: "Widow's Bite"}.Apply(ctx)
				},
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Widow's Bite"}.Apply(ctx)
				})
		},
	})
}
