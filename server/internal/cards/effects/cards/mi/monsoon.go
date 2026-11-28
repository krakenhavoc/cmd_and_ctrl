package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Monsoon — Enchantment {2}{R}{G} (Ice Age, S82 OLD, #5636):
//
//	"At the beginning of each player's end step, tap all untapped Islands that player
//	 controls and this enchantment deals X damage to the player, where X is the number
//	 of Islands tapped this way."
//
// Each player's end step, not just the controller's. Only Islands that were untapped
// and became tapped count (CR 701.26a), and the damage goes to that player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a86e7de8-8448-406c-b162-22c62bdc3813",
		Name:         "Monsoon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			}, "Monsoon — tap that player's untapped Islands; it deals X damage to them", effects.S82OldB841MonsoonTapIslands),
		},
	})
}
