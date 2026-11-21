package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invigorated Rampage — Instant {1}{R}:
//
//	"Choose one —
//	 • Target creature gets +4/+0 and gains trample until end of turn.
//	 • Two target creatures each get +2/+0 and gain trample until end of
//	   turn."
//
// Each bullet reads its own target group (ModeTargets); the second needs
// two distinct targets (CR 115.3). A target that has become illegal is
// skipped (CR 608.2b) and the other still gets its boost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1a665f1-b06f-4775-bcb1-c57a476fa651",
		Name:         "Invigorated Rampage",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gets +4/+0 and gains trample until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					return invigoratedRampagePump(ctx, occ, 4)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 4, 0, "trample"))),
			effects.ModeWithPurpose(effects.ModeDoing("Two target creatures each get +2/+0 and gain trample until end of turn.",
				effects.Distinct(effects.TargetCreature("two target creatures").WithCount(2, 2)),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					return invigoratedRampagePump(ctx, occ, 2)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "trample"))),
		),
	})
}

// invigoratedRampagePump gives each still-legal target of the mode
// occurrence +power/+0 and trample until end of turn.
func invigoratedRampagePump(ctx *effects.Context, occ, power int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (effects.BoostUntilEOT{Target: t.ID, Power: power, Label: "Invigorated Rampage — +power/+0"}).Apply(ctx); err != nil {
			return err
		}
		if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Invigorated Rampage — trample"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
