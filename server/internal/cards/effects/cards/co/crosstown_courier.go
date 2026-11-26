package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crosstown Courier — Creature — Vedalken {1}{U}, 2/1 (Return to Ravnica,
// S82 OLD):
//
//	"Whenever this creature deals combat damage to a player, that player
//	 mills that many cards."
//
// The damaged player and the amount are read as the trigger fires and carried
// on the stack item, so the mill still happens if the Courier has left by
// resolution. A library holding fewer cards mills what it has (CR 701.17b).
//
// No simplification.
func init() {
	const key = "Crosstown Courier — that player mills that many cards"
	effects.Register(effects.Spec{
		OracleID:     "82184f0a-54a1-4ae0-9b99-2b3450c6e544",
		Name:         "Crosstown Courier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: effects.ThisDealtCombatDamageToAPlayer,
			Key:       key,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, key)
				item.Params.Player = ev.Target
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.MillCards{Player: item.Params.Player, N: item.Params.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
