package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const theodenLabel = "Théoden, King of Rohan — target creature gains double strike until end of turn"

// Théoden, King of Rohan — Legendary Creature — Human Noble {1}{R}{W}, 2/3
// (The Lord of the Rings: Tales of Middle-earth Commander):
//
//	"Whenever Théoden or another Human you control enters, target creature
//	 gains double strike until end of turn."
//
// A "this or another Human" enters watch (Erkenbrand's shape) with a
// targeted keyword grant: any creature, not only the controller's. A target
// that left in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e739e3ce-4c5e-4172-a9de-adf66ca6588c",
		Name:         "Théoden, King of Rohan",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.On(game.EventETB, effects.GapB127ThisOrAnotherEntered("Human"), theodenLabel,
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"double strike"}, Label: theodenLabel}.Apply(ctx)
					}),
				effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "double strike"))),
		},
	})
}
