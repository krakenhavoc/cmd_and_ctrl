package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mystic Barrier — Enchantment {4}{W} (Commander 2013, S82 OLD b386):
//
//	"When this enchantment enters and at the beginning of your upkeep, choose
//	 left or right.
//	 Each player may attack only the nearest opponent in the last chosen
//	 direction and planeswalkers controlled by that opponent."
//
// Pramikon's restriction (ADR 0206), read off the permanent's chosen option,
// with the choice made again by a trigger each upkeep of its controller: the
// newest answer replaces the old one ("the last chosen direction"), and until
// the first answer lands the permanent restricts nothing. The trigger is one
// ability with two conditions (Obsessive Pursuit's shape).
//
// No simplification.
func init() {
	const name = "Mystic Barrier"
	choose := effects.ChooseLeftOrRightAsEnters(name)
	effects.Register(effects.Spec{
		OracleID:     "0caf42f5-abff-48aa-9bbf-df6cba169ef3",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, effects.ThisEntersOrYourUpkeep,
				"Mystic Barrier — choose left or right",
				func(g *game.Game, item *game.StackItem) error {
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					c.Controller = item.Controller
					return choose(&c, effects.NewContext(g, item))
				}),
		},
		AttackOnlyNearestOpponentInChosenDirection: true,
	})
}
