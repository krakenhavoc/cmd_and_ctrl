package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rivalry — Enchantment {2}{R}:
//
//	"At the beginning of each player's upkeep, if that player controls more
//	 lands than each other player, this enchantment deals 2 damage to them."
//
// The intervening "if" (CR 603.4) is asked as the upkeep begins and again as
// the ability resolves; "that player" is the player whose upkeep it is, not
// Rivalry's controller.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da77cf1b-eaeb-4a60-aa16-a1217ce8b550",
		Name:         "Rivalry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && effects.OldB718ControlsMoreLandsThanEachOtherPlayer(g, ev.Actor)
			}, "Rivalry — 2 damage to that player", func(g *game.Game, item *game.StackItem) error {
				p := effects.TriggeringActor(item)
				if p == uuid.Nil || !effects.OldB718ControlsMoreLandsThanEachOtherPlayer(g, p) {
					return nil
				}
				return effects.DealDamage{Source: item.SourceCardID, Target: p, Amount: 2}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
