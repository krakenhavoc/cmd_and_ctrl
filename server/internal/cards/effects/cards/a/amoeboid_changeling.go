package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Amoeboid Changeling — Creature — Shapeshifter {1}{U}, 1/1 (Lorwyn,
// #5593):
//
//	"Changeling (This card is every creature type.)
//	 {T}: Target creature gains all creature types until end of turn.
//	 {T}: Target creature loses all creature types until end of turn."
//
// Both rows are layer-4 type changes (CR 613.1d), not keyword grants:
// "gains all creature types" is Shields of Velis Vel's grant and "loses all
// creature types" is Nameless Inversion's, so the later of the two on one
// creature wins by timestamp, and a changeling that loses its types keeps
// the keyword but is no type until the turn ends (CR 514.2). Each needs the
// {T}, so the Changeling can't use either the turn it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "68865793-daeb-4c55-80cb-12853f9f96c4",
		Name:            "Amoeboid Changeling",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Target creature gains all creature types until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GrantAllCreatureTypesUntilEOT{Target: id, Label: "Amoeboid Changeling — all creature types"}.Apply(ctx)
				},
			},
			{
				Label:   "{T}: Target creature loses all creature types until end of turn.",
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
				Cost:    effects.TapCost(),
				Targets: effects.TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.UntilEndOfTurn(ctx, id, nil, "Amoeboid Changeling — loses all creature types",
						game.LoseAllCreatureTypesMod())
				},
			},
		},
	})
}
