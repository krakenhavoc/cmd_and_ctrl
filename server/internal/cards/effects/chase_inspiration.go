package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chase Inspiration — Instant {U}, Secrets of Strixhaven (#4456):
//
//	"Target creature you control gets +0/+3 and gains hexproof until end of
//	 turn. (It can't be the target of spells or abilities your opponents
//	 control.)"
//
// Both halves are keyed to the one announced target: a layer-7c +0/+3 and a
// layer-6 hexproof grant, each until end of turn. A target that left in
// response gets neither (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9c055ef0-1164-4a9f-80de-23405423aed5",
		Name:         "Chase Inspiration",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 0, 3, "hexproof")),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: t.ID, Toughness: 3, Label: "Chase Inspiration — +0/+3"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Chase Inspiration — hexproof"}.Apply(ctx)
		},
	})
}
