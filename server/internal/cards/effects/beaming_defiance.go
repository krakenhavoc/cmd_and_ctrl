package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beaming Defiance — Instant {1}{W} (Strixhaven: School of Mages, #5502):
//
//	"Target creature you control gets +2/+2 and gains hexproof until end
//	 of turn. (It can't be the target of spells or abilities your opponents
//	 control.)"
//
// Blossoming Defense's two primitives: the +2/+2 (layer 7c) and the
// hexproof (layer 6), on a creature you control, judged at announce and
// again as the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f88eeb04-4b8e-4c76-a8c6-90887b94d518",
		Name:         "Beaming Defiance",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2, "hexproof")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Beaming Defiance — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"hexproof"}, Label: "Beaming Defiance — hexproof"}.Apply(ctx)
		},
	})
}
