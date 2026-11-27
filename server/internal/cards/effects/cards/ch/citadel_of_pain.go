package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Citadel of Pain — Enchantment {2}{R} (Prophecy, S82 OLD b670, #5617):
//
//	"At the beginning of each player's end step, this enchantment deals X
//	 damage to that player, where X is the number of untapped lands they
//	 control."
//
// Blood Clock's each-player trigger: the player whose end step it is (the
// event's actor), not the Citadel's controller, takes the damage. X is read as
// the trigger resolves (CR 608.2h), counting that player's untapped lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2233c079-5ca6-4992-8bd2-84dbf9ed53a9",
		Name:         "Citadel of Pain",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AnyPlayer, "Citadel of Pain — it deals damage to that player equal to the number of untapped lands they control",
				func(g *game.Game, item *game.StackItem) error {
					player := item.Trigger.Event.Actor
					if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
						return nil
					}
					n := effects.CountControlled(g, player, func(c game.Card) bool { return c.IsLand() && !c.Tapped })
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
