package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:      "39f2a632-30d2-4b35-9e7f-fef27713a4f7",
		Name:          "Widow's Bite",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Teamwork(3)},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains deathtouch until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"deathtouch"}, Label: "Widow's Bite"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch"))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets -2/-2 until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Widow's Bite"}.Apply(ctx)
				}), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		).InsteadIf(2, effects.TeamworkUsed),
	})
}
