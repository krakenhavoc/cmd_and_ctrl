package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arborea Pegasus — Creature — Pegasus {3}{W}, 2/3 (Adventures in the Forgotten Realms):
//
//	Flying
//	 When this creature enters, target creature gets +1/+1 and gains flying until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a17452d-3adb-42ec-99f8-0083740cefa0",
		Name:            "Arborea Pegasus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Arborea Pegasus — target creature gets +1/+1 and gains flying until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range effects.LegalTargetCardIDs(ctx) {
					if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Arborea Pegasus — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Arborea Pegasus — flying"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying"))),
		},
	})
}
