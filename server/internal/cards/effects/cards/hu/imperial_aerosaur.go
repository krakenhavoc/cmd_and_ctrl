package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Imperial Aerosaur — Creature — Dinosaur {3}{W}, 3/3 (Jumpstart 2022, #5948):
//
//	"Flying
//	 When this creature enters, another target creature you control gets
//	 +1/+1 and gains flying until end of turn."
//
// Cunning Coyote's enters trigger with flying: "another" excludes the
// Aerosaur itself, so with no other creature the trigger is removed
// (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6889a555-6fde-4728-9cf2-715f0730efb7",
		Name:            "Imperial Aerosaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Imperial Aerosaur — another target creature you control gets +1/+1 and gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Imperial Aerosaur — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Imperial Aerosaur — flying"}.Apply(ctx)
				}),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying"))),
		},
	})
}
