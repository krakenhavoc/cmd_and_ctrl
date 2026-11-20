package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "68865793-daeb-4c55-80cb-12853f9f96c4",
		Name:            "Amoeboid Changeling",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordChangeling},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Target creature gains all creature types until end of turn.",
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return GrantAllCreatureTypesUntilEOT{Target: id, Label: "Amoeboid Changeling — all creature types"}.Apply(ctx)
				},
			},
			{
				Label:   "{T}: Target creature loses all creature types until end of turn.",
				Purpose: ForTargets(BenefitUntilEOT(0, 0, 0)),
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return untilEndOfTurn(ctx, id, nil, "Amoeboid Changeling — loses all creature types",
						game.LoseAllCreatureTypesMod())
				},
			},
		},
	})
}
