package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Built to Last — Instant {W} (Jumpstart 2022, #5948):
//
//	"Target creature gets +2/+2 until end of turn. If it's an artifact
//	 creature, it gains indestructible until end of turn."
//
// The artifact check reads the creature through the layered view as the
// spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "403e554b-72a5-48b3-acea-0b4062ffb913",
		Name:         "Built to Last",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Built to Last — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found || !c.IsArtifact() {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}}.Apply(ctx)
		},
	})
}
