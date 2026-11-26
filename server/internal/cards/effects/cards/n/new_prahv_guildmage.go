package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// New Prahv Guildmage — Creature — Human Wizard {W}{U}, 2/2 (Return to Ravnica,
// #5566):
//
//	"{W}{U}: Target creature gains flying until end of turn.
//	 {3}{W}{U}: Detain target nonland permanent an opponent controls."
//
// The detained permanent can't attack or block and its activated abilities can't
// be activated until your next turn (CR 701.35a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5912d4b7-d18d-4619-a96b-d2615200971f",
		Name:         "New Prahv Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}{U}: Target creature gains flying until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
				Cost:    effects.ManaCost("{W}{U}"),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "New Prahv Guildmage — flying until end of turn"}.Apply(ctx)
				},
			},
			{
				Label:   "{3}{W}{U}: Detain target nonland permanent an opponent controls.",
				Cost:    effects.ManaCost("{3}{W}{U}"),
				Targets: effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls()),
				Effect:  effects.OldB423DetainTargets("New Prahv Guildmage — detain"),
			},
		},
	})
}
