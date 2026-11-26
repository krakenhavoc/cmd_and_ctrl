package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wind Dancer — Creature — Faerie {1}{U}, 1/1 (Conspiracy, S82 OLD batch 363,
// #5557):
//
//	"Flying
//	 {T}: Target creature gains flying until end of turn."
//
// Flying is the printed keyword; the ability pins its target at resolution
// (CR 611.2c) and grants flying until end of turn. A target that left is
// skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9903d75-f71d-45d0-ac5f-955c0c227e02",
		Name:            "Wind Dancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Wind Dancer — flying"}.Apply(ctx)
			},
		}},
	})
}
