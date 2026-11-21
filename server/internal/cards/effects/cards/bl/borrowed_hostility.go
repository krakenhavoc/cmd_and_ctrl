package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borrowed Hostility — Instant {R}:
//
//	"Escalate {3} (Pay this cost for each mode chosen beyond the
//	 first.)
//	 Choose one or both —
//	 • Target creature gets +3/+0 until end of turn.
//	 • Target creature gains first strike until end of turn."
//
// Escalate (CR 702.120a, #2126): {3} for the second mode.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b45fc33b-0658-4fe1-89da-54a07a12ebd4",
		Name:         "Borrowed Hostility",
		Completeness: effects.CompletenessFull,
		Modes: effects.Escalating(effects.ChooseN("Choose one or both", 1, 2,
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +3/+0 until end of turn.",
				effects.TargetCreature("target creature"),
				effects.BoostTheModesTarget(3, 0, "Borrowed Hostility")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0))),
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains first strike until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Borrowed Hostility"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike"))),
		), effects.EscalateMana("{3}")),
	})
}
