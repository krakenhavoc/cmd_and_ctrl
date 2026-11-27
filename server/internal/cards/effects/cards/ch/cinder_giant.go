package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cinder Giant — {3}{R} Creature — Giant, 5/3 (Weatherlight, S82 OLD batch 771,
// #5629):
//
//	"At the beginning of your upkeep, this creature deals 2 damage to each other
//	 creature you control."
//
// At each of its controller's upkeeps, one damage event from the Giant to every
// OTHER creature that player controls as it resolves, not the Giant itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3a8306ae-3162-4105-95f4-dacd10c7c051",
		Name:         "Cinder Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Cinder Giant — it deals 2 damage to each other creature you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.DamageEachMatching(ctx, effects.And(effects.Creature(), effects.YouControl(), effects.OtherThan(item.SourceCardID)), 2)
				}),
		},
	})
}
