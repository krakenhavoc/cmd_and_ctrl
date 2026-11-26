package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predator, Flagship — Legendary Artifact {5} (Commander 2014):
//
//	"{2}: Target creature gains flying until end of turn.
//	 {5}, {T}: Destroy target creature with flying."
//
// Two activated abilities, each with its own target judged again as it
// resolves (CR 608.2b): a creature that has lost flying by then is not
// destroyed.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "647c1339-9508-4bb9-8e77-ebc3dfdf4069",
		Name:         "Predator, Flagship",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: Target creature gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying"))},
				Cost:    effects.ManaCost("{2}"),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Predator, Flagship — flying"}.Apply(ctx)
				},
			},
			{
				Label:   "{5}, {T}: Destroy target creature with flying.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
				Effect:  effects.DestroyFirstLegalTarget,
			},
		},
	})
}
