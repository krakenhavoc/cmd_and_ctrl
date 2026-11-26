package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Boom Keg — Artifact {4} (Fate Reforged):
//
//	"At the beginning of your upkeep, sacrifice this artifact.
//	 When this artifact is put into a graveyard from the battlefield, it deals 3
//	 damage to any target."
//
// Two triggers (CR 603.6c, 700.4): the sacrifice is only one of the ways the Keg
// can reach a graveyard, and any other (destroyed, countered as a copy, legend
// rule) pays off the same. The damage trigger looks back at the Keg as it left.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e819d68-3a6e-449a-8392-e072724b80f1",
		Name:         "Goblin Boom Keg",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Goblin Boom Keg — sacrifice this artifact", effects.SacrificeThisPermanentAtYourEndStep),
			effects.Targeting(effects.WhenThisDies("Goblin Boom Keg — it deals 3 damage to any target",
				func(g *game.Game, item *game.StackItem) error {
					for _, t := range effects.NewContext(g, item).LegalTargets() {
						return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: 3}.Apply(effects.NewContext(g, item))
					}
					return nil
				}), effects.TargetAny()),
		},
	})
}
