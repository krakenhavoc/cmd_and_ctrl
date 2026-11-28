package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Primal Order — Enchantment {2}{G}{G} (Homelands, S82 OLD b824, #5635):
//
//	"At the beginning of each player's upkeep, this enchantment deals damage to
//	that player equal to the number of nonbasic lands they control."
//
// Every upkeep, whoever's it is (AtEachUpkeep): the player is the event's
// actor, and the count is read as the trigger resolves, through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "63136274-df97-49ae-9efe-1385410e744b",
		Name:         "Primal Order",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Primal Order — it deals damage to that player equal to the number of nonbasic lands they control",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Trigger.Event.Actor
					if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
						return nil
					}
					ctx := effects.NewContext(g, item)
					n := len(effects.MatchingBattlefield(ctx, effects.And(effects.NonbasicLand(), effects.ControlledBy(player))))
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: n}.Apply(ctx)
				}),
		},
	})
}
