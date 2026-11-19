package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d1a665f1-b06f-4775-bcb1-c57a476fa651",
		Name:         "Invigorated Rampage",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Target creature gets +4/+0 and gains trample until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					return invigoratedRampagePump(ctx, occ, 4)
				}), ForTargets(BenefitUntilEOT(0, 4, 0, "trample"))),
			ModeWithPurpose(ModeDoing("Two target creatures each get +2/+0 and gain trample until end of turn.",
				Distinct(TargetCreature("two target creatures").WithCount(2, 2)),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					return invigoratedRampagePump(ctx, occ, 2)
				}), ForTargets(BenefitUntilEOT(0, 2, 0, "trample"))),
		),
	})
}

// invigoratedRampagePump gives each still-legal target of the mode
// occurrence +power/+0 and trample until end of turn.
func invigoratedRampagePump(ctx *Context, occ, power int) error {
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: power, Label: "Invigorated Rampage — +power/+0"}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Invigorated Rampage — trample"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
