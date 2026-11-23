package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flensing Raptor — Creature — Phyrexian Bird {2}{W}, 2/2
// (Phyrexia: All Will Be One, #5479):
//
//	"Flying
//	 Toxic 1
//	 When this creature enters, another target creature you control with
//	 toxic gets +1/+1 and gains flying until end of turn."
//
// With no other creature of yours that has toxic there is no target and
// the trigger is removed as it would go on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5d55c397-3828-4e4b-8a61-8c7272ba4800",
		Name:            "Flensing Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "toxic 1"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Flensing Raptor — another target creature you control with toxic gets +1/+1 and flying",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Flensing Raptor — +1/+1"}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Flensing Raptor — flying"}.Apply(ctx)
					}),
				effects.Another(effects.TargetCreature("another target creature you control with toxic", effects.YouControl(), effects.WithToxic()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "flying"))),
		},
	})
}
