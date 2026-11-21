package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pseudodragon Familiar — Creature — Dragon {2}{U}, 2/1 (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	"Flying
//	 {2}{U}: Target creature gains flying until end of turn."
//
// GrantKeywordUntilEOT on the one target, still legal as the ability resolves
// (CR 608.2b). Flying is a canonical keyword, so the grant is real.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de9cee6e-d95b-4e8b-99e4-3b43a0a542aa",
		Name:            "Pseudodragon Familiar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{2}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Pseudodragon Familiar — flying"}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
