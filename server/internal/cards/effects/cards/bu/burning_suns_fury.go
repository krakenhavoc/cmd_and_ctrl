package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Sun's Fury — Instant {1}{R} (March of the Machine): "Convoke. Up
// to two target creatures each get +2/+0 and gain haste until end of
// turn."
//
// Each target still legal as it resolves is affected (CR 608.2b); up to
// two is WithCount(0, 2) (CR 601.2c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c0073d7-a12a-45ba-8ac3-9e2230e2fb06",
		Name:         "Burning Sun's Fury",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0, "haste")),
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, id := range effects.LegalTargetCardIDs(ctx) {
				if err := (effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 0, Label: "Burning Sun's Fury — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Burning Sun's Fury — haste"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
