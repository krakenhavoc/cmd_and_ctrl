package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonerise Spirit — Creature — Spirit Bird {1}{W}, 1/2 (Strixhaven: School of
// Mages, #5502):
//
//	"Flying
//	 {4}, Exile a card from your graveyard: Target creature gains flying until
//	 end of turn."
//
// The exiled card is a cost chosen as the ability is activated (CR 602.2b); it
// is any card in the graveyard, never the Spirit itself (a battlefield
// permanent has none). The grant is Battle-Rage Blessing's keyword grant on the
// one target, skipped if the target is gone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9fbab5bc-c651-46dc-89b6-0bcfeadb3891",
		Name:            "Stonerise Spirit",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, Exile a card from your graveyard: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.ExileFromGraveyard(1, "a card", nil)),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Stonerise Spirit — flying until end of turn"}.Apply(ctx)
			},
		}},
	})
}
