package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roar of Challenge — Sorcery {2}{G} (Jumpstart 2022, #5948):
//
//	"All creatures able to block target creature this turn do so.
//	 Ferocious — That creature gains indestructible until end of turn if
//	 you control a creature with power 4 or greater."
//
// Bloodscent's lure, then the ferocious check as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9d83a1d3-ef0f-4134-a38b-c992d396592b",
		Name:         "Roar of Challenge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := effects.LureTargetForTheTurn(ctx, "Roar of Challenge — all creatures able to block it do so"); err != nil {
				return err
			}
			if !effects.FerociousYouControlPowerFourOrGreater(ctx) {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Roar of Challenge — indestructible"}.Apply(ctx)
		},
	})
}
