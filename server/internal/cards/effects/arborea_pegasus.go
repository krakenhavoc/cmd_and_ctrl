package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arborea Pegasus — Creature — Pegasus {3}{W}, 2/3 (Adventures in the Forgotten Realms):
//
//	Flying
//	 When this creature enters, target creature gets +1/+1 and gains flying until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9a17452d-3adb-42ec-99f8-0083740cefa0",
		Name:            "Arborea Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Arborea Pegasus — target creature gets +1/+1 and gains flying until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, id := range legalTargetCardIDs(ctx) {
					if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Arborea Pegasus — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Arborea Pegasus — flying"}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 1, 1, "flying"))),
		},
	})
}
