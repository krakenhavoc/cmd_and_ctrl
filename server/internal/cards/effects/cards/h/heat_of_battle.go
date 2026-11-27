package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heat of Battle — Enchantment {1}{R} (Stronghold, S82 OLD batch 747, #5627):
//
//	"Whenever a creature blocks, this enchantment deals 1 damage to that
//	 creature's controller."
//
// Carnage Gladiator's block trigger: one per blocking creature, whoever
// controls it (CR 509.3a: the first of its blocks, so a creature blocking
// several attackers triggers once), dealing the damage to the player who
// controls that blocker as it resolves, or as it last was if it has gone. The
// enchantment is the damage source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "400285bb-ded9-4018-a7ea-3e6f4d24b900",
		Name:         "Heat of Battle",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBlock && ev.Amount <= 1
			}, "Heat of Battle — 1 damage to that creature's controller",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					player := ev.Actor
					if c, ok := g.LookupCardForEffect(ev.CardID); ok {
						player = c.Controller
					}
					if p := g.PlayerByIDForEffect(player); p == nil || p.Eliminated {
						return nil
					}
					return effects.DealDamage{Source: item.SourceCardID, Target: player, Amount: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
