package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "9c055ef0-1164-4a9f-80de-23405423aed5",
		Name:         "Chase Inspiration",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 3, "hexproof")),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Toughness: 3, Label: "Chase Inspiration — +0/+3"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"hexproof"}, Label: "Chase Inspiration — hexproof"}.Apply(ctx)
		},
	})
}
