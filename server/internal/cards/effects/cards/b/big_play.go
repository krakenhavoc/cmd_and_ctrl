package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Big Play — {1}{G} Instant:
//
//	"Target creature gets +2/+2 and gains reach until end of turn. Put a
//	+1/+1 counter on it."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "607425ac-9e90-4c9e-88bf-72f8a76cce97",
		Name:         "Big Play",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "reach")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Big Play — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"reach"}, Label: "Big Play — reach"}).Apply(ctx); err != nil {
				return err
			}
			return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
		},
	})
}
