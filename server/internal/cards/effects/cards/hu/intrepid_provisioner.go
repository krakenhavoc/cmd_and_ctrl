package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intrepid Provisioner — Creature — Human Scout {3}{G}, 3/3 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Trample
//	 When this creature enters, another target Human you control gets
//	 +2/+2 until end of turn."
//
// The target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b); the Provisioner itself is not a legal target
// ("another"), and a Human that has left by resolution gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d84873a-c079-4a0f-8878-c62e4bade20a",
		Name:            "Intrepid Provisioner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Intrepid Provisioner — another target Human you control gets +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Intrepid Provisioner — +2/+2 until end of turn"}.Apply(ctx)
				}),
				effects.Another(effects.PermanentYouControl("another target Human you control", effects.HasSubtype("Human")))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2))),
		},
	})
}
