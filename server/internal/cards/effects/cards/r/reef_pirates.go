package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reef Pirates — Creature — Zombie Pirate {1}{U}{U}, 2/2 (Homelands, S82 OLD b824,
// #5635):
//
//	"Whenever this creature deals damage to an opponent, that player mills a
//	card."
//
// Any damage, not only combat damage: one trigger per damage event, milling the
// player who was dealt it (Mindcrank's read of the event's target).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13fb5413-0057-4022-84da-2c90ce065ed1",
		Name:         "Reef Pirates",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtDamageToAnOpponent,
				"Reef Pirates — that player mills a card",
				func(g *game.Game, item *game.StackItem) error {
					victim := item.Trigger.Event.Target
					if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
						return nil
					}
					return effects.MillCards{Player: victim, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
